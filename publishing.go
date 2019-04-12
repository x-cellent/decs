package decs

func (cbus *CommandBus) publish(cmd *command) {
	// there will be no error since the producer handles broken connections gracefully
	_ = cbus.producer.Publish(cmd.Name_, cmd)
}
