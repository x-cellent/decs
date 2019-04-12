package decs

func (cbus *CommandBus) SubscribeAfterSuccess(eventName string, callback func(data interface{}, dispatcher EventDispatcher)) {
	cbus.Subscribe(eventName, AfterSuccess, callback)
}

func (cbus *CommandBus) SubscribeAfterFailure(eventName string, callback func(data interface{}, dispatcher EventDispatcher)) {
	cbus.Subscribe(eventName, AfterFailure, callback)
}

func (cbus *CommandBus) SubscribeAfter(eventName string, callback func(data interface{}, dispatcher EventDispatcher)) {
	cbus.Subscribe(eventName, After, callback)
}

func (cbus *CommandBus) Subscribe(eventName string, trigger Trigger, callback func(data interface{}, dispatcher EventDispatcher)) {
	m := cbus.ebus.subs[eventName]
	if m == nil {
		m = make(map[Trigger][]func(interface{}, EventDispatcher))
		cbus.ebus.subs[eventName] = m
	}

	m[trigger] = append(m[trigger], callback)
}

func (cbus *CommandBus) SubscribeConditionally(condition func(eventName string) bool, trigger Trigger, callback func(data interface{}, dispatcher EventDispatcher)) {
	cbus.ebus.condSubs = append(cbus.ebus.condSubs, &conditionalCallback{
		condition: condition,
		trigger:   trigger,
		callback:  callback,
	})
}
