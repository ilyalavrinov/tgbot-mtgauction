package db

import (
	"fmt"

	"github.com/admirallarimda/tgbotbase"
)

var chatSubscribedKey = "mtgauction:subscribed"

type ChatDB struct {
	props tgbotbase.PropertyStorage
}

func NewChatDB(props tgbotbase.PropertyStorage) *ChatDB {
	return &ChatDB{props: props}
}

func (db *ChatDB) AddChat(chatId int64) {
	db.props.SetPropertyForChat(chatSubscribedKey, tgbotbase.ChatID(chatId), 1)
}

func (db *ChatDB) RemoveChat(chatId int64) {
	db.props.SetPropertyForChat(chatSubscribedKey, tgbotbase.ChatID(chatId), 0)
}

func (db *ChatDB) GetAll() (users, chats []int64, err error) {
	vals, err := db.props.GetEveryHavingProperty(chatSubscribedKey)
	if err != nil {
		err = fmt.Errorf("unable to get chat subscribed properties: %w", err)
		return
	}
	for _, v := range vals {
		if v.Value != "1" {
			continue
		}
		if v.Chat != tgbotbase.ChatID(0) {
			chats = append(chats, int64(v.Chat))
		} else if v.User != tgbotbase.UserID(0) {
			users = append(users, int64(v.User))
		}
	}
	return
}
