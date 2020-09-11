package topdeck

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/admirallarimda/tgbotbase"
	"github.com/go-redis/redis"
	"github.com/gocolly/colly"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/log"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type pollDB struct {
	conn redis.Conn
}

func NewPollDB(conn redis.Conn) *pollDB {
	return &pollDB{
		conn: conn,
	}
}

func (db *pollDB) toggleSeen(l lot) bool {
	// TODO: implement
	return true
}

type poller struct {
	cron   tgbotbase.Cron
	lotsCh chan []lot
	db     pollDB

	outMsgCh chan<- tgbotapi.Chattable
}

func NewPoller(cron tgbotbase.Cron) tgbotbase.BackgroundMessageHandler {
	p := &poller{
		cron:   cron,
		lotsCh: make(chan []lot),
	}
	return p
}

func (p *poller) Init(outMsgCh chan<- tgbotapi.Chattable, _ chan<- tgbotbase.ServiceMsg) {
	p.outMsgCh = outMsgCh
}

func (p *poller) Name() string {
	return "Topdeck poller"
}

func (p *poller) Run() {
	p.cron.AddJob(time.Now(), &pollerJob{lotsCh: p.lotsCh})
	go p.processLots()
}

func (p *poller) processLots() {
	for lots := range p.lotsCh {
		seenCnt := 0
		for _, l := range lots {
			seen := p.db.toggleSeen(l)
			if seen {
				seenCnt++
			} else {
				log.Debugw("Processing unseen lot",
					"id", l.ID,
					"name", l.Name)
			}
		}
		log.Debugw("Processed seen lots",
			"count", seenCnt)
	}
}

type pollerJob struct {
	lotsCh chan<- []lot
}

var re = regexp.MustCompile("JSON\\.parse\\(\\\"(.*)\\\"\\)\\,")

type lot struct {
	ID            string `json:"id"`
	Name          string `json:"lot"`
	Image         string `json:"image"`
	DatePublished string `json:"date_published"`
	DateEstimated string `json:"date_estimated"`

	BidInitial string `json:"start_bid"`
	BidCurrent string `json:"current_bid"`
	BidCount   string `json:"bid_amount"`
}

func (j *pollerJob) Do(scheduledWhen time.Time, cron tgbotbase.Cron) {
	defer cron.AddJob(time.Now().Add(10*time.Minute), j)

	var lots []lot
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

		dec := json.NewDecoder(strings.NewReader(result))
		_, err := dec.Token()
		if err != nil {
			log.Errorw("get opening failed",
				"err", err)
			return
		}
		for dec.More() {
			var l lot
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

	j.lotsCh <- lots
}
