package main

import (
	"flag"

	"github.com/admirallarimda/tgbotbase"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/bot"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/bot/db"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/log"
	"gopkg.in/gcfg.v1"
)

var argCfg = flag.String("cfg", "./bot.cfg", "path to config")

type config struct {
	tgbotbase.Config
	Redis tgbotbase.RedisConfig
}

func main() {
	flag.Parse()

	var cfg config

	if err := gcfg.ReadFileInto(&cfg, *argCfg); err != nil {
		log.Fatalw("Cannot read config file",
			"filename", *argCfg)
	}

	tgbot := tgbotbase.NewBot(tgbotbase.Config{TGBot: cfg.TGBot, Proxy_SOCKS5: cfg.Proxy_SOCKS5})

	cron := tgbotbase.NewCron()
	pool := tgbotbase.NewRedisPool(cfg.Redis)
	auctionDB := db.NewAuctionDB(pool.GetConnByName("mtgauction"))
	chatDB := db.NewChatDB(tgbotbase.NewRedisPropertyStorage(pool))
	tgbot.AddHandler(tgbotbase.NewBackgroundMessageDealer(bot.NewPollReceiver(cron, auctionDB, chatDB)))
	tgbot.AddHandler(tgbotbase.NewEngagementMessageDealer(bot.NewEngagementHandler(chatDB)))

	log.Info("Starting bot")
	tgbot.Start()
	log.Info("Stopping bot")
}
