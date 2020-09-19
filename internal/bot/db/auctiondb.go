package db

import (
	"fmt"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/ilyalavrinov/tgbot-mtgauction/internal/log"
)

type AuctionPlatform int

const (
	PlatformTopdeck  = iota
	PlatformMtgtrade = iota
)

func (p AuctionPlatform) String() string {
	res := "UNKNOWN"
	switch p {
	case PlatformTopdeck:
		res = "topdeck"
	case PlatformMtgtrade:
		res = "mtgtrade"
	}
	return res
}

type Auction struct {
	ID            string `json:"id" redis:"id"`
	Plaform       AuctionPlatform
	Name          string `json:"lot" redis:"lot"`
	Image         string `json:"image" redis:"image"`
	DatePublished string `json:"date_published" redis:"date_published"`
	DateEstimated string `json:"date_estimated" redis:"date_estimated"`

	BidInitial string `json:"start_bid" redis:"bid_initial"`
	BidCurrent string `json:"current_bid" redis:"bid_current"`
	BidCount   string `json:"bid_amount" redis:"bid_count"`
}

type AuctionDB struct {
	conn *redis.Client
}

func NewAuctionDB(conn *redis.Client) *AuctionDB {
	return &AuctionDB{
		conn: conn,
	}
}

func key(a Auction) string {
	return fmt.Sprintf("mtgauction:%s:%s", a.Plaform, a.ID)
}

func (db *AuctionDB) Accept(a Auction) (seen, dateChanged bool, err error) {
	res := db.conn.Exists(key(a))
	if res.Err() != nil {
		return false, false, fmt.Errorf("cannot check if record exists: %w", res.Err())
	}
	if res.Val() != 0 {
		seen = true
		dateChanged = false // TODO: placeholder for some distant future
		err = nil
		return
	}

	err = db.saveAuction(a)
	if err != nil {
		return false, false, fmt.Errorf("cannot save auction record: %w", res.Err())
	}

	return
}

func (db *AuctionDB) saveAuction(a Auction) error {
	v := reflect.ValueOf(a)
	t := reflect.TypeOf(a)
	fieldsN := v.NumField()
	savedValue := make(map[string]interface{}, fieldsN)
	for i := 0; i < fieldsN; i++ {
		ft := t.Field(i)
		name, ok := ft.Tag.Lookup("redis")
		if !ok {
			continue
		}
		savedValue[name] = v.Field(i).Interface()
	}
	k := key(a)
	status := db.conn.HMSet(k, savedValue)
	log.Debugw("saved to redis",
		"key", k,
		"fieldsN", len(savedValue),
		"err", status.Err(),
		"status", status.Val(),
	)
	return status.Err()
}
