package nats

import (
	"encoding/json"
	"log"


	"github.com/nats-io/nats.go"
)

// Publisher отвечает за публикацию событий в NATS.
type Publisher struct {
	nc *nats.Conn
}

// NewPublisher создаёт подключение к NATS по URL.
func NewPublisher(url string) (*Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	log.Printf("[nats] connected to %s", url)
	return &Publisher{nc: nc}, nil
}

// WalletUpdatedEvent структура события обновления баланса.
type WalletUpdatedEvent struct {
	UserId  string `json:"user_id"`
	Balance int32  `json:"balance"`
}

// PublishWalletUpdated публикует WalletUpdatedEvent в тему "wallet.updated".
func (p *Publisher) PublishWalletUpdated(userID string, balance int32) {
	evt := WalletUpdatedEvent{UserId: userID, Balance: balance}
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("[nats] marshal error: %v", err)
		return
	}
	if err := p.nc.Publish("wallet.updated", data); err != nil {
		log.Printf("[nats] publish error: %v", err)
	} else {
		log.Printf("[nats] published wallet.updated: %s", data)
	}
}
