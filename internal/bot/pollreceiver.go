package bot

import (
	"fmt"
	"time"

	"github.com/admirallarimda/tgbotbase"
	"github.com/admirallarimda/tgbotbase/pkg/tgbotutil"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/bot/db"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/bot/topdeck"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/log"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type pollReceiver struct {
	cron  tgbotbase.Cron
	lotCh chan db.Auction
	db    *db.AuctionDB
	chats *db.ChatDB

	outMsgCh chan<- tgbotapi.Chattable
}

func NewPollReceiver(cron tgbotbase.Cron, auctionDB *db.AuctionDB, chatsDB *db.ChatDB) tgbotbase.BackgroundMessageHandler {
	p := &pollReceiver{
		cron:  cron,
		lotCh: make(chan db.Auction),

		db:    auctionDB,
		chats: chatsDB,
	}
	return p
}

func (p *pollReceiver) Init(outMsgCh chan<- tgbotapi.Chattable, _ chan<- tgbotbase.ServiceMsg) {
	p.outMsgCh = outMsgCh
}

func (p *pollReceiver) Name() string {
	return "Poll Receiver"
}

func (p *pollReceiver) Run() {
	p.cron.AddJob(time.Now(), topdeck.NewPoller(p.lotCh))
	go p.processLots()
}

func (p *pollReceiver) processLots() {
	for lot := range p.lotCh {
		seen, _, err := p.db.Accept(lot)
		if err != nil {
			log.Errorw("error saving auction",
				"err", err)
		}
		if seen {
			continue
		}
		users, chats, err := p.chats.GetAll()
		if err != nil {
			log.Errorw("Cannot get a list of subscribers",
				"err", err)
			continue
		}
		log.Debugw("Processing unseen lot",
			"id", lot.ID,
			"name", lot.Name,
			"subscribersN", len(users)+len(chats))

		picFName, err := tgbotutil.LoadPicToTmp(lot.PicURL(), "mtgauction-")
		if err != nil {
			log.Errorw("failed loading pic",
				"url", lot.PicURL(),
				"err", err)
			continue
		}
		text := fmt.Sprintf("[%s](%s)\nCurrent bid: %s\nEnds at: %s", tgbotutil.EscapeMarkdown(lot.Name), lot.URL(), lot.BidCurrent, lot.DateEstimated)

		for _, chat := range chats {
			msg := tgbotapi.NewPhotoUpload(int64(chat), picFName)
			msg.Caption = text
			msg.ParseMode = "MarkdownV2"
			p.outMsgCh <- msg
		}
	}
}
