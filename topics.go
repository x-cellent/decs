package decs

func (cbus *CommandBus) Topics() []string {
	var tt []string
	for t := range cbus.defs {
		tt = append(tt, t)
	}
	return tt
}
