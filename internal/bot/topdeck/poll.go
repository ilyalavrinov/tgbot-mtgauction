package topdeck

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/admirallarimda/tgbotbase"
	"github.com/gocolly/colly"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/bot/db"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/log"
)

type pollerJob struct {
	lotCh chan<- db.Auction
}

func NewPoller(ch chan<- db.Auction) *pollerJob {
	return &pollerJob{lotCh: ch}
}

var re = regexp.MustCompile("JSON\\.parse\\(\\\"(.*)\\\"\\)\\,")

func (j *pollerJob) Do(scheduledWhen time.Time, cron tgbotbase.Cron) {
	defer cron.AddJob(time.Now().Add(10*time.Minute), j)

	var lots []db.Auction
	c := colly.NewCollector()

	c.OnHTML("script", func(e *colly.HTMLElement) {
		matches := re.FindAllSubmatch([]byte(e.Text), -1)
		if len(matches) == 0 {
			return
		}
		text := matches[0][1]
		pos := 0
		result := ""
		for pos < len(text) {
			if text[pos] == '\\' && text[pos+1] == 'u' {
				s := fmt.Sprintf("'%s'", text[pos:pos+6])
				c, err := strconv.Unquote(s)
				if err != nil {
					log.Errorw("Unquote failed",
						"err", err)
					return
				}
				result = result + c
				pos += 6
			} else {
				result = result + string(text[pos])
				pos++
			}
		}
		result = strings.ReplaceAll(result, "\\\"", "")
		result = strings.ReplaceAll(result, "\\", "")
		//fmt.Printf("RESULT %s\n", result)

		dec := json.NewDecoder(strings.NewReader(result))
		_, err := dec.Token()
		if err != nil {
			log.Errorw("get opening failed",
				"err", err)
			return
		}
		for dec.More() {
			var l db.Auction
			err := dec.Decode(&l)
			if err != nil {
				log.Errorw("decode failed",
					"err", err)
				time.Sleep(1 * time.Second)
				continue
			}

			lots = append(lots, l)
		}
		_, err = dec.Token()
		if err != nil {
			log.Errorw("read closing failed",
				"err", err)
			return
		}

		log.Debugw("Scraped lots",
			"count", len(lots))
	})

	url := "https://topdeck.ru/apps/toptrade/auctions"
	err := c.Visit(url)
	if err != nil {
		log.Errorw("Unable to scrape",
			"url", url,
			"err", err)
		return
	}

	for _, l := range lots {
		j.lotCh <- l
	}
}
