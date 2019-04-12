package mq

import (
	"sync"
	"time"
)

type Producer interface {
	Publish(topic string, data []byte) error
	CreateTopic(topic string) error
	CreateChannel(topic, channel string) error
}

type ProducerProxy struct {
	retryActive bool
	provider    Provider
	producer    Producer
	mtx         *sync.RWMutex
	q           []*pendingMessage
	err         error
	flushing    bool
}

type pendingMessage struct {
	topic string
	data  []byte
}

// NewProducerProxy creates a new producer proxy that will publish all messages through the
// given producer even if the connection breaks. If that happens all
// messages will be cached and eventually published in order as soon as the
// connection is established again.
// Each cached message will then be published after the given flush delay.
func NewProducerProxy(provider Provider, retryInterval, flushDelay time.Duration) *ProducerProxy {
	producer, _ := provider.CreateProducer()
	p := &ProducerProxy{
		retryActive: retryInterval > 0 && flushDelay >= 0,
		provider:    provider,
		producer:    producer,
		mtx:         &sync.RWMutex{},
	}

	if !p.retryActive {
		return p
	}

	if retryInterval <= 0 {
		retryInterval = DefaultRetryInterval
	}

	if flushDelay < 0 {
		flushDelay = DefaultFlushDelay
	}

	go func() {
		for {
			err := p.err
			for err != nil {
				p.producer, err = p.provider.CreateProducer()
				time.Sleep(retryInterval)
			}

			p.mtx.RLock()
			if len(p.q) > 0 {
				// try to publish oldest pending message
				msg := p.q[0]
				p.err = p.producer.Publish(msg.topic, msg.data)
				if p.err != nil {
					// connection still broken or broken during flushing
					p.flushing = false
					p.mtx.RUnlock()
					continue
				}

				// succeeded => initiate flushing (if not done already)
				// and remove pending message from queue
				p.flushing = true
				p.mtx.RUnlock()
				p.mtx.Lock()
				p.q = p.q[1:]
				p.mtx.Unlock()
				time.Sleep(flushDelay)
				continue
			}

			p.flushing = false
			p.mtx.RUnlock()
			time.Sleep(retryInterval)
		}
	}()

	return p
}

func (p *ProducerProxy) Publish(topic string, message []byte) error {
	if !p.retryActive {
		return p.producer.Publish(topic, message)
	}

	if p.flushing || p.err != nil || p.publish(topic, message) != nil {
		// flushing or connection still broken => enqueue pending message
		p.mtx.Lock()
		p.q = append(p.q, &pendingMessage{
			topic: topic,
			data:  message,
		})
		p.mtx.Unlock()
	}

	return nil
}

func (p *ProducerProxy) publish(topic string, message []byte) error {
	p.err = p.producer.Publish(topic, message)
	return p.err
}

func (p *ProducerProxy) CreateTopic(topic string) error {
	return p.producer.CreateTopic(topic)
}

func (p *ProducerProxy) CreateChannel(topic, channel string) error {
	return p.producer.CreateChannel(topic, channel)
}
