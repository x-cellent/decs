package nats

import (
	"github.com/nats-io/go-nats"
	"github.com/x-cellent/decs/mq"
	"go.uber.org/zap"
)

type provider struct {
	*config
	consumer *nats.Conn
	log      *zap.Logger
}

func NewProvider(log *zap.Logger, url string) mq.Provider {
	if len(url) == 0 {
		return nil
	}

	cfg := newConfig(url)

	if log == nil {
		log = zap.NewNop()
	}

	return &provider{
		config:   cfg,
		consumer: newConsumer(log, url),
		log:      log,
	}
}

func (p *provider) CreateProducer() (mq.Producer, error) {
	return newProducer(p.config)
}

func (p *provider) CreateConsumer(topic, channel string, callback func([]byte) error) {
	_, err := p.consumer.Subscribe(topic, func(msg *nats.Msg) {
		err := callback(msg.Data)
		if err != nil {
			p.log.Sugar().Error("Failed to process command", "topic", topic, "error", err)
		}
	})
	if err != nil {
		p.log.Sugar().Error("Failed to create consumer", "topic", topic, "error", err)
	}
}
