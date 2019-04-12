package nsq

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

type consumerRegistration struct {
	consumer *consumer
	log      *zap.Logger
	c        *nsq.Consumer
}

type consumer struct {
	lookupdHttpAddresses []string
	config               *nsq.Config
	log                  *zap.Logger
}

func newConsumer(log *zap.Logger, lookupdHttpAddresses ...string) *consumer {
	cfg := nsq.NewConfig()
	_ = cfg.Set("lookupd_poll_interval", "5s")
	_ = cfg.Set("default_requeue_delay", "5s")

	return &consumer{
		config:               cfg,
		lookupdHttpAddresses: lookupdHttpAddresses,
		log:                  log,
	}
}

func (c *consumer) MustRegister(topic, channel string) *consumerRegistration {
	cr, err := c.Register(topic, channel)
	if err != nil {
		panic(err)
	}
	return cr
}

func (c *consumer) Register(topic, channel string) (*consumerRegistration, error) {
	C, err := nsq.NewConsumer(topic, channel, c.config)
	if err != nil {
		return nil, fmt.Errorf("cannot create consumer for topic:%q, channel:%q: %v", topic, channel, err)
	}
	return &consumerRegistration{
		consumer: c,
		log:      c.log,
		c:        C,
	}, nil
}

func (cr *consumerRegistration) Consume(callback func([]byte) error, concurrent int) error {
	cr.c.SetLogger(cr, nsq.LogLevelError)
	cr.c.AddConcurrentHandlers(nsq.HandlerFunc(func(nsqmsg *nsq.Message) error {
		if callback == nil {
			return nil
		}

		return callback(nsqmsg.Body)
	}), concurrent)

	return cr.c.ConnectToNSQLookupds(cr.consumer.lookupdHttpAddresses)
}

func (cr *consumerRegistration) Output(num int, msg string) error {
	cr.log.Error(msg)
	return nil
}
