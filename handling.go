package decs

import "sync/atomic"

func (cbus *CommandBus) handleCommands() {
	for {
		select {
		case cmd := <-cbus.queue.q:
			cbus.handleCommand(cmd)
			atomic.AddInt32(&cbus.incomingCount, -1)
		}
	}
}

func (cbus *CommandBus) handleCommand(cmd *command) {
	if cmd.Undo {
		cbus.undo(1)
		return
	}

	// ensure idempotence
	if cbus.versionTracker.alreadyHandled(cmd) {
		return
	}

	// custom filtering
	if cmd.bus.filterOut(cmd) {
		return
	}

	if cbus.cachingSuspended || cmd.Oneshot {
		cbus.doEventually(cmd, false)
		return
	}

	if cbus.isInFlow(cmd.Name_) {
		cbus.doEventually(cmd, true)

		cbus.flows[cmd.Name_] = append(cbus.flows[cmd.Name_], cmd)
		return
	}

	// record command if necessary
	if cbus.Recorder.IsRecording() {
		cbus.Recorder.append(cmd)
	}

	if cbus.defs[cmd.Name_].Flowing() {
		cbus.EnterFlow(cmd.Name_)
		defer cbus.ExitFlow(cmd.Name_)
	}

	cbus.doEventually(cmd, true)
}
