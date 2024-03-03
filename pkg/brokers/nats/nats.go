package nats

import (
	"fmt"

	"gogogo/config"

	"github.com/nats-io/nats.go"
)

type Broker struct {
	Conn *nats.Conn
}

func New(cfg *config.Config) (Broker, error) {
	nc, err := nats.Connect(cfg.BrokerDSN)
	if err != nil {
		return Broker{}, fmt.Errorf("nats connect: %w", err)
	}

	return Broker{Conn: nc}, nil
}
