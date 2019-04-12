package decs

import (
	"fmt"
	"sync/atomic"
)

func (cbus *CommandBus) createConsumer(topic string) {
	cbus.createTopic(topic)
	cbus.createChannel(topic, cbus.commandChannel(topic))
	cbus.createConsumerProxy(topic)
}

func (cbus *CommandBus) createTopic(topic string) {
	err := cbus.producer.CreateTopic(topic)
	if err != nil {
		cbus.log.Sugar().Error("Failed to create topic", "topic", topic, "error", err)
	}
}

func (cbus *CommandBus) createChannel(topic string, channel string) {
	err := cbus.producer.CreateChannel(topic, channel)
	if err != nil {
		cbus.log.Sugar().Error("Failed to create channel", "topic", topic, "channel", channel, "error", err)
	}
}

func (cbus *CommandBus) createConsumerProxy(topic string) {
	callback := func(cmd *command) error {
		if cmd.bus.filterLocalsOut(cmd) {
			return nil
		}

		atomic.AddInt32(&cbus.incomingCount, 1)
		cbus.queue.queue(cmd)

		return nil
	}

	cbus.mqProvider.CreateConsumer(topic, cbus.commandChannel(topic), cbus.callbackProxy(callback))
}

func (cbus *CommandBus) commandChannel(topic string) string {
	return fmt.Sprintf("%v-%v", topic, cbus.id)
}

func (cbus *CommandBus) callbackProxy(callback func(*command) error) func(bytes []byte) error {
	return func(bytes []byte) error {
		cmd := &command{bus: cbus}
		err := cmd.UnmarshalJSON(bytes)
		if err != nil {
			cbus.log.Sugar().Error(err)
			return err
		}

		return callback(cmd)
	}
}
