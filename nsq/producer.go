package nsq

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type producer struct {
	*nsq.Producer
	nsqdTcpAddress        string
	nsqLookupdHttpAddress string
	retryInterval         time.Duration
	log                   *zap.Logger
}

func newProducer(log *zap.Logger, nsqdTcpAddress, nsqLookupdHttpAddress string) (*producer, error) {
	cfg := nsq.NewConfig()
	nsqProducer, err := nsq.NewProducer(nsqdTcpAddress, cfg)
	if err != nil {
		return nil, err
	}

	producer := &producer{
		Producer:              nsqProducer,
		nsqdTcpAddress:        nsqdTcpAddress,
		nsqLookupdHttpAddress: nsqLookupdHttpAddress,
		log:                   log,
	}

	nsqProducer.SetLogger(producer, nsq.LogLevelError)

	return producer, nil
}

func (p *producer) Publish(topic string, data []byte) error {
	return p.Producer.Publish(topic, data)
}

func (p *producer) CreateTopic(topic string) error {
	rsp, err := http.Post(fmt.Sprintf("http://%s/topic/create?topic=%v", p.nsqLookupdHttpAddress, topic), "text/plain", nil)
	if err != nil {
		p.log.Sugar().Error("Failed to create topic", "topic", topic)
		return err
	}
	_ = rsp.Body.Close()
	return nil
}

func (p *producer) CreateChannel(topic, channel string) error {
	rsp, err := http.Post(fmt.Sprintf("http://%s/channel/create?topic=%v&channel=%v", p.nsqLookupdHttpAddress, topic, channel), "text/plain", nil)
	if err != nil {
		p.log.Sugar().Error("Failed to create channel", "topic", topic, "channel", channel)
		return err
	}
	_ = rsp.Body.Close()
	return nil
}

func (p *producer) Output(num int, msg string) error {
	p.log.Error(msg)
	return nil
}
