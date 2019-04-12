package decs

import (
	"github.com/x-cellent/decs/mq"
	"github.com/x-cellent/decs/nats"
	"github.com/x-cellent/decs/nats/stan"
	"github.com/x-cellent/decs/nsq"
	"go.uber.org/zap"
	"time"
)

func (cbus *CommandBus) Logger() *zap.Logger {
	return cbus.log
}

func (cbus *CommandBus) SetLogger(logger *zap.Logger) {
	cbus.log = logger
	cbus.Recorder.log = logger
}

func (cbus *CommandBus) SetMqProvider(provider mq.Provider, retryInterval, flushDelay time.Duration) {
	if cbus.mqProvider != nil && cbus.mqProvider == provider {
		return
	}

	local := provider == nil
	if local {
		provider = mq.NewLocalProvider()
	}

	cbus.mqProvider = provider

	if local {
		cbus.producer = newProducerProxy(nil, cbus)
	} else {
		cbus.producer = newProducerProxy(mq.NewProducerProxy(cbus.mqProvider, retryInterval, flushDelay), cbus)
	}
}

func (cbus *CommandBus) ConfigureLocalProvider(retryInterval, flushDelay time.Duration) {
	cbus.SetMqProvider(nil, retryInterval, flushDelay)
}

func (cbus *CommandBus) ConfigureNatsProvider(retryInterval, flushDelay time.Duration, url string) {
	cbus.SetMqProvider(nats.NewProvider(cbus.log, url), retryInterval, flushDelay)
}

func (cbus *CommandBus) ConfigureNatsStreamingProvider(retryInterval, flushDelay time.Duration, url, clusterID, clientID string) {
	cbus.SetMqProvider(stan.NewProvider(cbus.log, url, clusterID, clientID), retryInterval, flushDelay)
}

func (cbus *CommandBus) ConfigureNsqProvider(retryInterval, flushDelay time.Duration, nsqdTcpAddress, nsqdHttpAddress string, nsqLookupdHttpAddresses ...string) {
	cbus.SetMqProvider(nsq.NewProvider(cbus.log, nsqdTcpAddress, nsqdHttpAddress, nsqLookupdHttpAddresses...), retryInterval, flushDelay)
}

func (cbus *CommandBus) HandleLocalCommandsImmediately() {
	if cbus.localFilter != nil {
		return
	}
	cbus.localFilter = newLocalCommandFilter(cbus)
}

func (cbus *CommandBus) HandleLocalCommandsOnDemand() {
	cbus.localFilter = nil
}

func (cbus *CommandBus) SuspendPublishing() {
	cbus.publishingSuspended = true
}

func (cbus *CommandBus) ResumePublishing() {
	cbus.publishingSuspended = false
}

func (cbus *CommandBus) SuspendCaching() {
	cbus.cachingSuspended = true
}

func (cbus *CommandBus) ResumeCaching() {
	cbus.cachingSuspended = false
}
