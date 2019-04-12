package decs

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const test = "test"

func TestRecorderAndPlayer(t *testing.T) {
	// given
	cbus := NewDefaultCommandBus(InMemory)
	cr := &CommandDefinition{
		Name:     test,
		DataType: &userData{},
	}
	cbus.DefineCommand(cr)

	cbus.RegisterCommandHandler(test, func(cmd Command, delegate Delegate, notifier ResultNotifier) {
		fmt.Println(time.Now().String())
	})
	cbus.RegisterUndoHandler(test, func(cmd Command, delegate Delegate) {
		fmt.Printf("Undo at time %v\n", time.Now().String())
	})

	// when
	cbus.Recorder.Start()
	cnt := 4
	for i := 0; i < cnt; i++ {
		cbus.Do(cbus.NewCommand(test, nil))
	}
	time.Sleep(time.Duration(cnt) * time.Millisecond) // Not needed when commands are handled immediately
	cbus.Recorder.Stop()
	cbus.Player.replay = cbus.Recorder.record

	require.Equal(t, cnt, len(cbus.Player.replay.Commands))

	cbus.UndoAll()

	cbus.Player.Start()
	for !cbus.Player.HasFinished() {
		time.Sleep(interval)
	}
}
