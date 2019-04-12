package mq

import (
	"time"
)

const (
	DefaultRetryInterval = 30 * time.Second
	DefaultFlushDelay    = 20 * time.Millisecond
)

type Provider interface {
	CreateProducer() (Producer, error)
	CreateConsumer(topic, channel string, callback func([]byte) error)
}

type localProvider struct{}

func NewLocalProvider() *localProvider {
	return &localProvider{}
}

func (p *localProvider) CreateProducer() (Producer, error) {
	return nil, nil
}

func (p *localProvider) CreateConsumer(topic, channel string, callback func([]byte) error) {
}
