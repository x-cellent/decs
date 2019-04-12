package decs

func (cbus *CommandBus) Reset() {
	cbus.Clear()
	cbus.Filters = cbus.Filters[:0]
	cbus.defs = make(map[string]*CommandDefinition)
	cbus.guards = make(map[string][]Guard)
	cbus.handlers = make(map[string]func(Command, Delegate, ResultNotifier))
	cbus.undoHandlers = make(map[string]func(Command, Delegate))

}

func (cbus *CommandBus) Clear() {
	cbus.publishingSuspended = false
	cbus.cachingSuspended = false
	cbus.undoables = cbus.undoables[:0]
	cbus.redoables = cbus.redoables[:0]
	cbus.flows = make(map[string][]*command)
	cbus.Recorder.record.Clear()
	cbus.store.Clear()
}
