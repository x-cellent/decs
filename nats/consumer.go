package nats

import (
	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

func newConsumer(log *zap.Logger, url string) *nats.Conn {
	//cfg := newConfig(url) //TODO Make connection configurable, e.g. securing the connection. Do this also for stan and nsq
	nc, err := nats.Connect(url)
	if err != nil {
		log.Sugar().Error("Cannot create NATS consumer", "error", err)
	}

	return nc
}
