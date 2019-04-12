package decs

func (cbus *CommandBus) RedoLast() bool {
	if len(cbus.redoables) == 0 {
		return false
	}

	cbus.Redo(1)
	return true
}

func (cbus *CommandBus) RedoAll() {
	cbus.Redo(len(cbus.redoables))
}

func (cbus *CommandBus) Redo(redoCount int) {
	var from int
	if len(cbus.redoables) <= redoCount {
		from = 0
	} else {
		from = len(cbus.redoables) - redoCount
	}
	rr := cbus.redoables[from:]
	cbus.redoables = cbus.redoables[:from]

	for i := len(rr) - 1; i >= 0; i-- {
		// redo it
		cbus.do(rr[i])
	}
}
