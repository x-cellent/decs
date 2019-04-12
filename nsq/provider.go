package nsq

import (
	"github.com/x-cellent/decs/mq"
	"go.uber.org/zap"
)

type provider struct {
	*config
	consumer *consumer
	log      *zap.Logger
}

func NewProvider(log *zap.Logger, nsqdTcpAddress, nsqdHttpAddress string, nsqLookupdHttpAddresses ...string) mq.Provider {
	if len(nsqdTcpAddress) == 0 || len(nsqdHttpAddress) == 0 || len(nsqLookupdHttpAddresses) == 0 || len(nsqLookupdHttpAddresses[0]) == 0 {
		return nil
	}

	if log == nil {
		log = zap.NewNop()
	}

	return &provider{
		config:   newConfig(nsqdTcpAddress, nsqdHttpAddress, nsqLookupdHttpAddresses...),
		consumer: newConsumer(log, nsqLookupdHttpAddresses...),
		log:      log,
	}
}

func (p *provider) CreateProducer() (mq.Producer, error) {
	return newProducer(p.log, p.nsqdTcpAddress, p.nsqdHttpAddress)
}

func (p *provider) CreateConsumer(topic, channel string, callback func([]byte) error) {
	cr := p.consumer.MustRegister(topic, channel)
	err := cr.Consume(callback, 1)
	if err != nil {
		p.log.Sugar().Error("Failed to create consumer", "topic", topic, "channel", channel, "error", err)
	}
}
