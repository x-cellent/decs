package nats

import (
	"github.com/nats-io/go-nats"
)

type producer struct {
	*nats.Conn
	config *config
}

func newProducer(cfg *config) (*producer, error) {
	conn, err := nats.Connect(cfg.url)
	if err != nil {
		return nil, err
	}

	return &producer{
		Conn:   conn,
		config: cfg,
	}, nil
}

func (p *producer) Publish(topic string, data []byte) error {
	return p.Conn.Publish(topic, data)
}

func (p *producer) CreateTopic(topic string) error {
	return nil
}

func (p *producer) CreateChannel(topic, channel string) error {
	return nil
}
