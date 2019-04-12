package decs

import (
	"encoding/json"
	"fmt"
	"github.com/x-cellent/decs/mq"
	"sync/atomic"
)

type producerProxy struct {
	mq.Producer
	cbus *CommandBus
}

func newProducerProxy(producer mq.Producer, cbus *CommandBus) *producerProxy {
	return &producerProxy{
		Producer: producer,
		cbus:     cbus,
	}
}

func (p *producerProxy) Publish(topic string, data interface{}) error {
	cmd := data.(*command)

	bb, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("cannot marshal data to json: %v", err)
	}

	if !cmd.internal && !cmd.Oneshot && !p.publishingSuspended() {
		_ = p.Producer.Publish(topic, bb)

		// defer command handling if localFilter is not active
		if p.cbus.localFilter == nil {
			return nil
		}
	}

	cmd, err = cmd.bus.UnmarshalCommand(bb)
	if err != nil {
		return fmt.Errorf("cannot unmarshal command from json: %v", err)
	}

	if atomic.LoadInt32(&cmd.bus.incomingCount) > 0 {
		cmd.bus.queue.queue(cmd)
		return nil
	}

	cmd.bus.handleCommand(cmd)

	return nil
}

func (p *producerProxy) CreateTopic(topic string) error {
	if !p.publishingSuspended() {
		return p.Producer.CreateTopic(topic)
	}

	return nil
}

func (p *producerProxy) CreateChannel(topic, channel string) error {
	if !p.publishingSuspended() {
		return p.Producer.CreateChannel(topic, channel)
	}

	return nil
}

func (p *producerProxy) publishingSuspended() bool {
	return p == nil || p.Producer == nil || p.cbus.publishingSuspended
}
