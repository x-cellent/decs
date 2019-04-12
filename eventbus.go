package decs

import (
	"time"
)

type ResultNotifier interface {
	DispatchSuccess(event Event)
	DispatchFailure(event Event)
	NotifySuccess(eventName string, data interface{})
	NotifyFailure(eventName string, data interface{})
}

type EventDispatcher interface {
	Dispatch(event Event)
}

type conditionalCallback struct {
	condition func(string) bool
	trigger   Trigger
	callback  func(interface{}, EventDispatcher)
}

// Implements both, ResultNotifier and EventDispatcher
type eventBus struct {
	queue      []*event
	errQueue   []*event
	subs       map[string]map[Trigger][]func(interface{}, EventDispatcher)
	condSubs   []*conditionalCallback
	commandBus *CommandBus
	causation  Message
}

func newEventBus() *eventBus {
	return &eventBus{
		subs: make(map[string]map[Trigger][]func(interface{}, EventDispatcher)),
	}
}

func (ebus *eventBus) DispatchSuccess(evt Event) {
	event := evt.(*event)
	event.trigger = AfterSuccess
	ebus.queue = append(ebus.queue, event)
}

func (ebus *eventBus) DispatchFailure(evt Event) {
	event := evt.(*event)
	event.trigger = AfterFailure
	ebus.errQueue = append(ebus.errQueue, event)
}

func (ebus *eventBus) NotifySuccess(eventName string, data interface{}) {
	ebus.DispatchSuccess(NewEvent(eventName, data))
}

func (ebus *eventBus) NotifyFailure(eventName string, data interface{}) {
	ebus.DispatchFailure(NewEvent(eventName, data))
}

// Dispatches all events synchronously
func (ebus *eventBus) dispatchAll() {
	queue := make([]*event, len(ebus.queue))
	copy(queue, ebus.queue)
	ebus.queue = ebus.queue[:0]

	errQueue := make([]*event, len(ebus.errQueue))
	copy(errQueue, ebus.errQueue)
	ebus.errQueue = ebus.errQueue[:0]

	for _, evt := range queue {
		ebus.Dispatch(evt)
	}

	for _, evt := range errQueue {
		ebus.Dispatch(evt)
	}
}

// Dispatches event synchronously
func (ebus *eventBus) Dispatch(evt Event) {
	event, ok := evt.(*event)
	if !ok {
		event = castEvent(evt)
	}

	event.Id = ebus.commandBus.idGenerator.Create()
	event.version = ebus.commandBus.versionTracker.next()
	event.raised = time.Now()
	if ebus.causation != nil {
		event.aggregateID = ebus.causation.AggregateID()
		event.correlationID = ebus.causation.CorrelationID()
		event.causationID = ebus.causation.ID()
	}
	ebus.causation = event

	m := ebus.subs[event.name]
	if m != nil {
		for _, callback := range m[event.trigger] {
			callback(event.data, ebus)
		}
		for _, callback := range m[After] {
			callback(event.data, ebus)
		}
	}

	for _, cc := range ebus.condSubs {
		if cc.condition(event.name) {
			if cc.trigger != After && cc.trigger != event.trigger {
				continue
			}
			cc.callback(event.data, ebus)
		}
	}
}
