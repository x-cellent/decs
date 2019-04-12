package decs

import (
	"github.com/satori/go.uuid"
	"time"
)

//TODO type Command struct {...}
type Command interface {
	Name() string
	Category() string
	Data() interface{}
	ID() string
	AggregateID() string
	CorrelationID() string
	CausationID() string
	Version() uint
	Raised() time.Time
	Handled() time.Time
}

type command struct {
	Name_       string      `json:"name"`
	Cat         string      `json:"category"`
	Data_       interface{} `json:"data"`
	Id          string      `json:"id"`
	AggregateId string      `json:"aggID"`
	Version_    uint        `json:"version"`
	BusId       string      `json:"cbusID"`
	Raised_     time.Time   `json:"raised"`
	Undo        bool        `json:"undo"`
	Oneshot     bool        `json:"oneshot"` // oneshot commands are not cached and therewith cannot be undone
	numBytes    int
	handled     time.Time
	internal    bool // internal commands will never be published and are always locally handled
	bus         *CommandBus
}

func (cbus *CommandBus) NewCommand(name string, data interface{}) *command {
	return cbus.newCommand(name, data, false, false)
}

func (cbus *CommandBus) NewInternalCommand(name string, data interface{}) *command {
	return cbus.newCommand(name, data, true, false)
}

func (cbus *CommandBus) NewOneshotCommand(name string, data interface{}) *command {
	return cbus.newCommand(name, data, false, true)
}

func (cbus *CommandBus) NewInternalOneshotCommand(name string, data interface{}) *command {
	return cbus.newCommand(name, data, true, true)
}

func (cbus *CommandBus) newCommand(name string, data interface{}, internal, oneshot bool) *command {
	return &command{
		Name_:       name,
		Cat:         cbus.defs[name].Category,
		Data_:       data,
		Id:          cbus.idGenerator.Create(),
		AggregateId: uuid.NewV4().String(),
		bus:         cbus,
		BusId:       cbus.id,
		Oneshot:     oneshot,
		internal:    internal,
	}
}

func (c *command) Name() string {
	return c.Name_
}

func (c *command) Category() string {
	return c.Cat
}

func (c *command) Data() interface{} {
	return c.Data_
}

func (c *command) ID() string {
	return c.Id
}

func (c *command) AggregateID() string {
	return c.AggregateId
}

func (c *command) CorrelationID() string {
	return c.Id
}

func (c *command) CausationID() string {
	return c.Id
}

func (c *command) Version() uint {
	return c.Version_
}

func (c *command) Raised() time.Time {
	return c.Raised_
}

func (c *command) Handled() time.Time {
	return c.handled
}

func (c *command) isLocal() bool {
	return c.BusId == c.bus.id
}
