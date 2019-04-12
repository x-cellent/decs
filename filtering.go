package decs

import "errors"

type Filter Guard

type localCommandFilter struct {
	cbusID string
}

func newLocalCommandFilter(cbus *CommandBus) *localCommandFilter {
	return &localCommandFilter{
		cbusID: cbus.id,
	}
}

func (f *localCommandFilter) Inspect(cmd *command) error {
	if cmd.isLocal() {
		return errors.New("local command")
	}
	return nil
}

func (cbus *CommandBus) filterLocalsOut(cmd *command) bool {
	return cbus.localFilter != nil && cbus.localFilter.Inspect(cmd) != nil
}

func (cbus *CommandBus) filterOut(cmd *command) bool {
	for _, f := range cbus.Filters {
		if err := f.Inspect(cmd); err != nil {
			cbus.log.Sugar().Info("Filtered out incoming command", "command", cmd, "reason", err)
			return true
		}
	}

	return false
}
