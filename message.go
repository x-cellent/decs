package decs

import "time"

// Commands and events are messages
type Message interface {
	ID() string
	AggregateID() string
	CorrelationID() string
	CausationID() string
	Version() uint
	Raised() time.Time
}
