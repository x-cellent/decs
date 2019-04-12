package decs

type Guard interface {
	Inspect(command Command) error
}

func (cbus *CommandBus) AddGlobalGuard(guard Guard) {
	cbus.AddCommandGuard("", guard)
}

func (cbus *CommandBus) AddCommandGuard(commandName string, guard Guard) {
	if len(commandName) == 0 {
		if !containsGuard(cbus.globalGuards, guard) {
			cbus.globalGuards = append(cbus.globalGuards, guard)
		}
		return
	}

	gg := cbus.guards[commandName]
	if gg == nil {
		cbus.guards[commandName] = []Guard{guard}
		return
	}

	if !containsGuard(gg, guard) {
		gg = append(gg, guard)
	}
}

func (cbus *CommandBus) RemoveGlobalGuard(guard Guard) {
	cbus.RemoveCommandGuard("", guard)
}

func (cbus *CommandBus) RemoveCommandGuard(commandName string, guard Guard) {
	if len(commandName) == 0 {
		removeGuard(cbus.globalGuards, guard)
		return
	}
	gg := cbus.guards[commandName]
	if gg != nil {
		removeGuard(gg, guard)
	}
}

func index(guards []Guard, guard Guard) int {
	for i, g := range guards {
		if g == guard {
			return i
		}
	}
	return -1
}

func containsGuard(guards []Guard, guard Guard) bool {
	return index(guards, guard) > -1
}

func removeGuard(guards []Guard, guard Guard) {
	i := index(guards, guard)
	if i > -1 {
		guards = append(guards[:i], guards[i+1:]...)
	}
}
