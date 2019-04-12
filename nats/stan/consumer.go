package stan

import (
	"github.com/nats-io/go-nats-streaming"
	"go.uber.org/zap"
)

func newConsumer(log *zap.Logger, url, clusterID, clientID string) stan.Conn {
	//cfg := newConfig(url, clusterID, clientID)
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Sugar().Error("Connection lost", "reason", reason)
		}))
	if err != nil {
		log.Sugar().Error("Cannot create NATS streaming consumer", "url", url, "clusterID", clusterID, "clientID", clientID, "error", err)
	}

	return sc
}
