package decs

import "fmt"

func (cbus *CommandBus) RegisterCommandHandler(commandName string, handler func(cmd Command, delegate Delegate, notifier ResultNotifier)) {
	if _, ok := cbus.defs[commandName]; !ok {
		panic(fmt.Errorf("before registering a handler for command %q you have to register the command itself", commandName))
	}
	if _, ok := cbus.handlers[commandName]; ok {
		panic(fmt.Errorf("handler for command %q already registered", commandName))
	}
	cbus.handlers[commandName] = handler
}

func (cbus *CommandBus) RegisterUndoHandler(commandName string, undoHandler func(cmd Command, delegate Delegate)) {
	if _, ok := cbus.defs[commandName]; !ok {
		panic(fmt.Errorf("before registering a handler for command %q you have to register the command itself", commandName))
	}
	if _, ok := cbus.undoHandlers[commandName]; ok {
		panic(fmt.Errorf("undo handler for command %q already registered", commandName))
	}
	cbus.undoHandlers[commandName] = undoHandler
}

func (cbus *CommandBus) RegisterDelegationHandler(handlerID string, delegations ...func(cmd Command, delegate Delegate, notifier ResultNotifier)) {
	cbus.delegate.delegations[handlerID] = append(cbus.delegate.delegations[handlerID], delegations...)
}

func (cbus *CommandBus) RegisterUndoDelegationHandler(handlerID string, delegations ...func(cmd Command, delegate Delegate)) {
	cbus.delegate.undoDelegations[handlerID] = append(cbus.delegate.undoDelegations[handlerID], delegations...)
}
