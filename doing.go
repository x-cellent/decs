package decs

import (
	"fmt"
	"reflect"
	"time"
)

func (cbus *CommandBus) Do(cmd Command) {
	if cmd == nil {
		return
	}

	c, ok := cmd.(*command)
	if !ok {
		// wrap custom commands, i.e. those that are not produced by any of decs.NewXXXCommand() constructors
		c = cbus.NewCommand(cmd.Name(), cmd.Data())
		c.AggregateId = cmd.AggregateID()
	}

	// verify that given command data is of expected type
	if c.Data_ != nil {
		dataType := reflect.TypeOf(c.Data_)
		expectedType := reflect.TypeOf(cbus.defs[c.Name_].DataType)
		if dataType != expectedType {
			cbus.log.Sugar().Error("Unexpected command data type", "command", c.Name_, "expected", expectedType.String(), "actual", dataType.String())
			return
		}
	}

	// reject command execution if at least one command guard vetoes
	for _, g := range cbus.guards[c.Name_] {
		if err := g.Inspect(c); err != nil {
			cbus.log.Sugar().Info("Handler guard prevents command from execution", "command", c, "guard", g, "error", err)
			return
		}
	}

	// reject command execution if at least one guard vetoes
	for _, g := range cbus.globalGuards {
		if err := g.Inspect(c); err != nil {
			cbus.log.Sugar().Info("Command guard prevents command from execution", "command", c, "guard", g, "error", err)
			return
		}
	}

	c.Version_ = cbus.versionTracker.next()
	c.Raised_ = time.Now()

	cbus.do(c)
}

func (cbus *CommandBus) do(cmd *command) {
	cmd.Undo = false
	cbus.publish(cmd)
}

func (cbus *CommandBus) doEventually(cmd *command, cache bool) {
	doHandler := cbus.handlers[cmd.Name_]
	if doHandler == nil {
		panic(fmt.Errorf("no handler for command %q registered", cmd.Name_))
	}

	cbus.delegate.currCmd = cmd
	cbus.delegate.undo = false

	doHandler(cmd, cbus.delegate, cbus.ebus)

	cmd.handled = time.Now()

	// append command to store
	if cache {
		cbus.store.push(cmd)
	}

	cbus.ebus.causation = cmd
	cbus.ebus.dispatchAll()

	// mark command as undoable
	if !cmd.Oneshot && !cbus.cachingSuspended {
		cat := cmd.Cat
		if len(cat) == 0 {
			cat = "..."
		}
		cbus.undoables = append(cbus.undoables, cmd)
		limit := cbus.defs[cmd.Name_].UndoableLimit
		if limit > UnlimitedUndoables && uint(len(cbus.undoables)) > limit {
			cbus.undoables = cbus.undoables[:limit]
		}
	}
}
