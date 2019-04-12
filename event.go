package decs

import "time"

const (
	PurgeApplication = "decs.purgeApplication"
)

type Trigger uint8

const (
	AfterSuccess Trigger = iota
	AfterFailure
	After
)

type Event interface {
	Name() string
	Data() interface{}
	Trigger() Trigger
	ID() string
	AggregateID() string
	CorrelationID() string
	CausationID() string
	Version() uint
	Raised() time.Time
}

type event struct {
	name    string
	trigger Trigger
	data    interface{}

	Id            string
	aggregateID   string
	correlationID string
	causationID   string
	version       uint
	raised        time.Time
}

func NewEvent(name string, data interface{}) Event {
	return &event{
		name:    name,
		trigger: After,
		data:    data,
	}
}

func castEvent(evt Event) *event {
	return &event{
		name:          evt.Name(),
		trigger:       evt.Trigger(),
		data:          evt.Data(),
		Id:            evt.ID(),
		aggregateID:   evt.AggregateID(),
		correlationID: evt.CorrelationID(),
		causationID:   evt.CausationID(),
	}
}

func (evt *event) Name() string {
	return evt.name
}

func (evt *event) Data() interface{} {
	return evt.data
}

func (evt *event) Trigger() Trigger {
	return evt.Trigger()
}

func (evt *event) ID() string {
	return evt.Id
}

func (evt *event) AggregateID() string {
	return evt.aggregateID
}

func (evt *event) CorrelationID() string {
	return evt.correlationID
}

func (evt *event) CausationID() string {
	return evt.causationID
}

func (evt *event) Version() uint {
	return evt.version
}

func (evt *event) Raised() time.Time {
	return evt.raised
}
