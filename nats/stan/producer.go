package stan

import (
	"github.com/nats-io/go-nats-streaming"
)

type producer struct {
	stan.Conn
	config *config
}

func newProducer(cfg *config) (*producer, error) {
	conn, err := stan.Connect(cfg.clusterID, cfg.clientID, stan.NatsURL(cfg.url))
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
