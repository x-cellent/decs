package decs

func (cbus *CommandBus) EnterFlow(commandName string) {
	if cbus.isInFlow(commandName) {
		return
	}

	cbus.flows[commandName] = []*command{}
}

func (cbus *CommandBus) ExitFlow(commandName string) {
	if !cbus.isInFlow(commandName) {
		return
	}

	delete(cbus.flows, commandName)
}

func (cbus *CommandBus) isInFlow(cmdName string) bool {
	return cbus.flows[cmdName] != nil
}

func getFlowCount(cbus *CommandBus, undoables []*command) int {
	l := len(undoables)
	if l == 0 {
		return 0
	}

	peek := undoables[l-1]

	cnt := 1
	for i := l - 2; i >= 0; i-- {
		if !cbus.defs[peek.Name()].RelatedInFlow(peek, undoables[i]) {
			break
		}
		cnt++
	}

	return cnt
}
