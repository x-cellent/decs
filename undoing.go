package decs

func (cbus *CommandBus) UndoLast() bool {
	l := len(cbus.undoables)
	if l == 0 {
		return false
	}

	cnt := 1
	if cbus.defs[cbus.undoables[l-1].Name_].Flowing() {
		cnt = getFlowCount(cbus, cbus.undoables)
	}

	cbus.publishUndoables(cnt)
	return true
}

func (cbus *CommandBus) UndoAll() {
	cbus.publishUndoables(-1)
}

func (cbus *CommandBus) publishUndoables(cnt int) {
	l := len(cbus.undoables)

	if cnt < 0 || cnt > l {
		cnt = l
	}
	var cc []*command
	var cmd *command
	for i := l - 1; i >= l-cnt; i-- {
		cmd = cbus.undoables[i]

		cmd.Undo = true
		cc = append(cc, cmd)

		// mark it as redoable
		cbus.redoables = append(cbus.redoables, cmd)
	}

	for _, cmd := range cc {
		cbus.publish(cmd)
	}
}

func (cbus *CommandBus) undo(undoCount int) {
	var c *command
	for i := 0; i < undoCount; i++ {
		if len(cbus.undoables) == 0 {
			return
		}

		// remove last undoable command
		c, cbus.undoables = cbus.undoables[len(cbus.undoables)-1], cbus.undoables[:len(cbus.undoables)-1]

		// undo all trailing events up to the one to be undone (including) and remove them from store
		var cmd *command
		for {
			// remove it from store
			cmd = cbus.store.delete(c)
			if cmd == nil {
				break
			}

			// undo trailing command
			uu := cbus.flows[cmd.Name_]
			var u *command
			for len(uu) > 0 {
				u, uu = uu[len(uu)-1], uu[:len(uu)-1]
				u.undoEventually(cbus)
			}

			cmd.undoEventually(cbus)

			if cmd.Id == c.Id {
				break
			}
		}

		// remove it from the record if command bus is recording
		rec := cbus.Recorder
		if rec.IsRecording() {
			_, rec.record.Commands = rec.record.Commands[len(rec.record.Commands)-1], rec.record.Commands[:len(rec.record.Commands)-1]
		}
	}
}

func (c *command) undoEventually(cbus *CommandBus) {
	def := cbus.defs[c.Name_]

	undoHandler := cbus.undoHandlers[c.Name_]
	if undoHandler != nil {
		cbus.delegate.currCmd = c
		cbus.delegate.undo = true

		undoHandler(c, cbus.delegate)
		cbus.versionTracker.update(cbus.id, c.Version_-1)
	} else {
		// TODO Consider to subscribe a PurgeApplication listener
		cbus.ebus.Dispatch(NewEvent(PurgeApplication, nil)) //TODO This is an ugly system event. Try to find a better solution
		cbus.versionTracker.update(cbus.id, 0)

		//TODO Introduce snapshots, i.e. run every once in a while "cbus.ebus.dispatch(CreateSnapshot, ".snapshot-3")"
		//     to let the user do the snapshot. Then undoing always from the last snapshot, i.e. user has to purge application
		//     and then load this last snapshot through "cbus.ebus.dispatch(LoadSnapshot, ".snapshot-3")" invoked right after
		//     the line above (purge application). This will increase performance dramatically, especially when already requested
		//     1.000.0000 commands or more. If the undo reach beyond the last snapshot use the one before, but do not delete it
		//     because we will need it if we redo right after. Nevertheless, it will be overwritten by the next upcoming snapshot.

		// redo all commands from scratch up to the current one (excluding) that effectively undoes it
		cs := cbus.store
		cs.moveSwap()
		cbus.store = NewCommandStore(cs.cbus, cs.memoryLimit)
		cs._apply(func(cmd *command) bool {
			if c.Id == cmd.Id {
				return false
			}
			cbus.doEventually(cmd, true)
			return true
		})
		cs._clear()
	}

	for _, evt := range def.UndoneEvents {
		cbus.ebus.NotifySuccess(evt, c.Data_)
	}

	cbus.ebus.causation = c
	cbus.ebus.dispatchAll()
}
