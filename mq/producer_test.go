package mq

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type provider struct {
	producer *producer
}

func (p *provider) CreateProducer() (Producer, error) {
	p.producer = &producer{}
	return p.producer, nil
}

func (p *provider) CreateConsumer(topic, channel string, callback func([]byte) error) {
}

type producer struct {
	fail      bool
	published [][]byte
}

func (p *producer) Publish(topic string, data []byte) error {
	if p.fail {
		return errors.New(topic)
	}
	p.published = append(p.published, data)
	return nil
}

func (p *producer) CreateTopic(topic string) error {
	return nil
}

func (p *producer) CreateChannel(topic, channel string) error {
	return nil
}

func TestProducerReconnection(t *testing.T) {
	// given
	p := &provider{}
	retryInterval := 100 * time.Millisecond
	flushDelay := time.Millisecond
	proxy := NewProducerProxy(p, retryInterval, flushDelay)
	require.True(t, proxy.retryActive)
	topic := "test"
	msgCount := 100
	var messages []string
	for i := 0; i < msgCount; i++ {
		messages = append(messages, fmt.Sprintf("Message %d", i))
	}

	// when
	p.producer.fail = true
	to := msgCount / 2
	for i := 0; i < to; i++ {
		_ = proxy.Publish(topic, []byte(messages[i]))
		require.NotNil(t, proxy.err)
	}

	// then
	require.Equal(t, 0, len(p.producer.published))
	require.False(t, proxy.flushing)

	// when
	p.producer.fail = false
	go func() {
		for {
			if proxy.flushing {
				break
			}
			time.Sleep(time.Nanosecond)
		}
		require.Nil(t, proxy.err)
		for i := to; i < len(messages); i++ {
			_ = proxy.Publish(topic, []byte(messages[i]))
			// ensure that at least one message has been published during flushing
			if i == to {
				require.True(t, proxy.flushing)
			}
		}
	}()
	time.Sleep(retryInterval)
	time.Sleep(time.Duration(3*msgCount) * flushDelay)

	// then
	require.Equal(t, len(messages), len(p.producer.published))
	for i := 0; i < len(messages); i++ {
		require.Equal(t, messages[i], string(p.producer.published[i]))
	}
}
