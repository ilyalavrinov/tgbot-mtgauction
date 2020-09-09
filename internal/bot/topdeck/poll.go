package topdeck

import (
	"github.com/admirallarimda/tgbotbase"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type poller struct {
}

func NewPoller() tgbotbase.BackgroundMessageHandler {
	p := &poller{}
	return p
}

func (p *poller) Init(chan<- tgbotapi.Chattable, chan<- tgbotbase.ServiceMsg) {

}

func (p *poller) Name() string {
	return "Topdeck poller"
}

func (p *poller) Run() {

}
