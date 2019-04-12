package nats

import (
	"crypto/tls"
)

type config struct {
	*tls.Config
	url string
}

func newConfig(url string) *config {
	return &config{
		Config: &tls.Config{},
		url:    url,
	}
}
