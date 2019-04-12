package decs

import "github.com/rs/xid"

type IDGenerator interface {
	Create() string
}

type defaultIdGenerator struct{}

func NewDefaultIDGenerator() *defaultIdGenerator {
	return &defaultIdGenerator{}
}

func (id *defaultIdGenerator) Create() string {
	return xid.New().String()
}
