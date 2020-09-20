package bot

import (
	"github.com/admirallarimda/tgbotbase"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/bot/db"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type engagementHandler struct {
	chats *db.ChatDB
}

func NewEngagementHandler(chats *db.ChatDB) tgbotbase.EngagementHandler {
	return &engagementHandler{chats: chats}
}

func (h *engagementHandler) Name() string {
	return "Engagement Handler"
}

func (h *engagementHandler) Engaged(chat *tgbotapi.Chat, user *tgbotapi.User) {
	if chat != nil {
		h.chats.AddChat(chat.ID)
	}
}

func (h *engagementHandler) Disengaged(chat *tgbotapi.Chat, user *tgbotapi.User) {
	if chat != nil {
		h.chats.RemoveChat(chat.ID)
	}
}
