package decs

import (
	"fmt"
	"strings"
)

const UnlimitedUndoables uint = 0

type CommandDefinition struct {
	Name          string
	DataType      interface{}
	RelatedInFlow func(lhs Command, rhs Command) bool
	UndoneEvents  []string
	AutoFocusable bool
	Category      string
	UndoableLimit uint
}

func (cd *CommandDefinition) Flowing() bool {
	return cd.RelatedInFlow != nil
}

func (cd *CommandDefinition) validate() {
	if cd == nil || len(strings.TrimSpace(cd.Name)) == 0 {
		panic(fmt.Errorf("invalid command definition: %+v", cd))
	}
}

func (cbus *CommandBus) DefineCommands(commandDefinitions ...*CommandDefinition) {
	for _, h := range commandDefinitions {
		cbus.DefineCommand(h)
	}
}

func (cbus *CommandBus) DefineCommand(commandDefinition *CommandDefinition) {
	if cbus.mqProvider == nil {
		panic("Please configure MQ provider first through CommandBus.SetMqProvider")
	}

	commandDefinition.validate()

	if _, ok := cbus.defs[commandDefinition.Name]; !ok {
		cbus.createConsumer(commandDefinition.Name)
	}

	cbus.defs[commandDefinition.Name] = commandDefinition
}

func (cbus *CommandBus) CanUndo(commandDefinition *CommandDefinition) bool {
	return cbus.undoHandlers[commandDefinition.Name] != nil
}
