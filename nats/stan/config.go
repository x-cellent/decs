package stan

import (
	"crypto/tls"
)

type config struct {
	*tls.Config
	url       string
	clusterID string
	clientID  string
}

func newConfig(url, clusterID, clientID string) *config {
	return &config{
		Config:    &tls.Config{},
		url:       url,
		clusterID: clusterID,
		clientID:  clientID,
	}
}
