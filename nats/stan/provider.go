package stan

import (
	"github.com/nats-io/go-nats-streaming"
	"github.com/x-cellent/decs/mq"
	"go.uber.org/zap"
	"time"
)

type provider struct {
	*config
	consumer stan.Conn
	log      *zap.Logger
}

//TODO Make all NATS streaming configuration customizable such as file or DB storage
func NewProvider(log *zap.Logger, url, clusterID, clientID string) mq.Provider {
	if len(clusterID) == 0 || len(clientID) == 0 {
		return nil
	}

	cfg := newConfig(url, clusterID, clientID)

	if log == nil {
		log = zap.NewNop()
	}

	return &provider{
		config:   cfg,
		consumer: newConsumer(log, url, clusterID, clientID),
		log:      log,
	}
}

func (p *provider) CreateProducer() (mq.Producer, error) {
	cfg := newConfig(p.url, p.clusterID, p.clientID+"-pub")
	return newProducer(cfg)
}

func (p *provider) CreateConsumer(topic, channel string, callback func([]byte) error) {
	_, err := p.consumer.Subscribe(
		topic,
		func(msg *stan.Msg) {
			err := msg.Ack()
			if err != nil {
				p.log.Sugar().Error("Failed to acknowledge message", "topic", topic, "error", err)
				return
			}

			err = callback(msg.Data)
			if err != nil {
				p.log.Sugar().Error("Failed to process command", "topic", topic, "error", err)
			}
		},
		stan.DurableName(channel),
		stan.MaxInflight(1), //TODO make this configurable by user
		stan.SetManualAckMode(),
		stan.AckWait(2*time.Second), //TODO make this configurable by user
		stan.DeliverAllAvailable(),  //TODO make this configurable by user
	)
	if err != nil {
		p.log.Sugar().Error("Failed to create NATS streaming consumer", "topic", topic, "error", err)
	}
}

/*
	// Simple Async Subscriber
	sub1, _ := sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})

	// Subscribe starting with most recently published value
	sub2, _ := sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	}, stan.StartWithLastReceived())

	// Receive all stored values in order
	sub3, _ := sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	}, stan.DeliverAllAvailable())

	// Receive messages starting at a specific sequence number
	sub4, _ := sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	}, stan.StartAtSequence(22))

	// Subscribe starting at a specific time
	var startTime time.Time
	sub5, _ := sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	}, stan.StartAtTime(startTime))

	// Subscribe starting a specific amount of time in the past (e.g. 30 seconds ago)
	sub6, _ := sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	}, stan.StartAtTimeDelta(30*time.Second))

	// Unsubscribe
	_ = sub1.Unsubscribe()
	_ = sub2.Unsubscribe()
	_ = sub3.Unsubscribe()
	_ = sub4.Unsubscribe()
	_ = sub5.Unsubscribe()
	_ = sub6.Unsubscribe()

	// Close connection
	_ = sc.Close()

*/
