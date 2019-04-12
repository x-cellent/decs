package decs

type Delegate interface {
	To(handlers ...string)
}

type delegate struct {
	ebus            *eventBus
	delegations     map[string][]func(cmd Command, delegate Delegate, notifier ResultNotifier)
	undoDelegations map[string][]func(cmd Command, delegate Delegate)
	undo            bool
	currCmd         *command
}

func newDelegate(eventBus *eventBus) *delegate {
	return &delegate{
		ebus:            eventBus,
		delegations:     make(map[string][]func(cmd Command, delegate Delegate, notifier ResultNotifier)),
		undoDelegations: make(map[string][]func(cmd Command, delegate Delegate)),
	}
}

func (d *delegate) To(handlers ...string) {
	if d.undo {
		var undoDelegations []func(cmd Command, delegate Delegate)
		for _, handler := range handlers {
			for _, undoDelegation := range d.undoDelegations[handler] {
				undoDelegations = append(undoDelegations, undoDelegation)
			}
		}

		for _, undoDelegation := range undoDelegations {
			undoDelegation(d.currCmd, d)
		}

		return
	}

	var delegations []func(cmd Command, delegate Delegate, notifier ResultNotifier)
	for _, handler := range handlers {
		for _, delegation := range d.delegations[handler] {
			delegations = append(delegations, delegation)
		}
	}

	for _, delegation := range delegations {
		delegation(d.currCmd, d, d.ebus)
	}
}
