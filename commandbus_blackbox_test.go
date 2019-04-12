package decs_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/x-cellent/decs"
	"github.com/x-cellent/decs/nats"
	"github.com/x-cellent/decs/nats/stan"
	"github.com/x-cellent/decs/nsq"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

const (
	AddOne               = "add-one"
	AddOneDelegation1    = "1-add-one-delegation"
	AddOneDelegation2    = "2-add-one-delegation"
	AddOneUndoDelegation = "add-one-undo-delegation"
	AddTwo               = "add-two"
	SubtractThree        = "subtract-three"
	MultiplyByFive       = "multiply-by-five"
	MultiplyByZero       = "multiply-by-zero"

	OnAddedOne         = "on-added-one"
	OnAddedTwo         = "on-added-two"
	OnSubtractedThree  = "on-subtracted-three"
	OnMultipliedByFive = "on-multiplied-by-five"
	OnMultipliedByZero = "on-multiplied-by-zero"

	OnAddOneUndone         = "on-add-one-UNDONE"
	OnAddTwoUndone         = "on-add-two-UNDONE"
	OnSubtractThreeUndone  = "on-subtract-three-UNDONE"
	OnMultiplyByFiveUndone = "on-multiply-by-five-UNDONE"
	OnMultiplyByZeroUndone = "on-multiply-by-zero-UNDONE"

	OnUpdate = "on-update"
)

type result struct {
	NewValue int `json:"value"`
	Value    int `json:"oldValue"`
}

var (
	wg     sync.WaitGroup
	mqProv string
)

// https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859
func prepare(t *testing.T, mqProvider string) func() {
	_ = os.Remove(decs.CommandStoreFile)

	var dockerCompose string
	var err error

	mqProv = mqProvider

	if len(mqProvider) > 0 {
		dockerCompose, err = exec.LookPath("docker-compose")
		if err != nil {
			t.FailNow()
		}

		env := ""
		for _, e := range os.Environ() {
			env += fmt.Sprintf("%s\n", e)
		}
		_ = ioutil.WriteFile("testdata/.env", []byte(env), 0644)

		err = exec.Command(dockerCompose, "-f", fmt.Sprintf("testdata/docker-compose.%s.yaml", mqProvider), "up", "-d").Run()
		if err != nil {
			t.FailNow()
		}
		time.Sleep(time.Second)
	}

	return cleanup
}

func cleanup() {
	if len(mqProv) > 0 {
		dockerCompose, err := exec.LookPath("docker-compose")
		if err == nil {
			_ = exec.Command(dockerCompose, "-f", fmt.Sprintf("testdata/docker-compose.%s.yaml", mqProv), "down").Run()
		}
	}
	_ = os.Remove(decs.CommandStoreFile)
	_ = os.Remove("testdata/.env")
}

func TestEventBusWithNsq(t *testing.T) {
	defer prepare(t, "nsq")()

	mqProvider := nsq.NewProvider(nil, "10.10.10.1:4150", "10.10.10.1:4151", "10.10.10.1:4161")
	cbus := decs.NewDefaultDistributedCommandBus(mqProvider, 0, 0, decs.KiloByte)

	runTest(t, cbus)
}

func TestEventBusWithNsqWithoutUndoHandlers(t *testing.T) {
	defer prepare(t, "nsq")()

	mqProvider := nsq.NewProvider(nil, "10.10.10.1:4150", "10.10.10.1:4151", "10.10.10.1:4161")
	cbus := decs.NewDefaultDistributedCommandBus(mqProvider, 0, 0, decs.KiloByte)

	runTestWithoutUndoHandlers(t, cbus)
}

func TestEventBusWithNats(t *testing.T) {
	port := "4222"
	_ = os.Setenv("HOST_PORT", port)
	defer prepare(t, "nats")()

	mqProvider := nats.NewProvider(nil, fmt.Sprintf("10.10.10.1:%s", port))
	cbus := decs.NewDefaultDistributedCommandBus(mqProvider, 0, 0, decs.InMemory)

	runTest(t, cbus)

	_ = os.Unsetenv("HOST_PORT")
}

func TestEventBusWithNatsWithoutUndoHandlers(t *testing.T) {
	port := "4223"
	_ = os.Setenv("HOST_PORT", port)
	defer prepare(t, "nats")()

	mqProvider := nats.NewProvider(nil, fmt.Sprintf("10.10.10.1:%s", port))
	cbus := decs.NewDefaultDistributedCommandBus(mqProvider, 0, 0, decs.InMemory)

	runTestWithoutUndoHandlers(t, cbus)

	_ = os.Unsetenv("HOST_PORT")
}

func TestEventBusWithNatsStreaming(t *testing.T) {
	defer prepare(t, "stan")()

	mqProvider := stan.NewProvider(nil, "10.10.10.1:4222", "decs", "client")
	cbus := decs.NewDefaultDistributedCommandBus(mqProvider, 0, 0, 183*decs.Byte)

	runTest(t, cbus)
}

func TestEventBusWithNatsStreamingWithoutUndoHandlers(t *testing.T) {
	defer prepare(t, "stan")()

	mqProvider := stan.NewProvider(nil, "10.10.10.1:4222", "decs", "client")
	cbus := decs.NewDefaultDistributedCommandBus(mqProvider, 0, 0, 183*decs.Byte)

	runTestWithoutUndoHandlers(t, cbus)
}

func TestUnconnectedEventBus(t *testing.T) {
	defer prepare(t, "")()

	cbus := decs.NewDefaultCommandBus(decs.Byte)

	runTest(t, cbus)
}

func TestUnconnectedEventBusWithoutUndoHandlers(t *testing.T) {
	defer prepare(t, "")()

	cbus := decs.NewDefaultCommandBus(decs.Byte)

	runTestWithoutUndoHandlers(t, cbus)
}

func runTest(t *testing.T, cbus *decs.CommandBus) {
	// given
	defs := []*decs.CommandDefinition{
		{
			Name:         AddOne,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnAddOneUndone},
		},
		{
			Name:         AddTwo,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnAddTwoUndone},
		},
		{
			Name:         MultiplyByFive,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnMultiplyByFiveUndone},
		},
		{
			Name:         SubtractThree,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnSubtractThreeUndone},
		},
		{
			Name:         MultiplyByZero,
			DataType:     &result{},
			UndoneEvents: []string{OnMultiplyByZeroUndone},
		},
	}
	cbus.DefineCommands(defs...)

	cbus.RegisterDelegationHandler(AddOneDelegation1, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value + 1
	})

	cbus.RegisterDelegationHandler(AddOneDelegation2, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		require.Equal(t, r.Value+1, r.NewValue)
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
	})

	cbus.RegisterCommandHandler(AddOne, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		delegate.To(AddOneDelegation1, AddOneDelegation2)
		notifier.NotifySuccess(OnAddedOne, r)
	})
	cbus.RegisterCommandHandler(AddTwo, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value + 2
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
		notifier.NotifySuccess(OnAddedTwo, r)
	})
	cbus.RegisterCommandHandler(MultiplyByFive, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value * 5
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
		notifier.NotifySuccess(OnMultipliedByFive, r)
	})
	cbus.RegisterCommandHandler(SubtractThree, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value - 3
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
		notifier.NotifySuccess(OnSubtractedThree, r)
	})
	cbus.RegisterCommandHandler(MultiplyByZero, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value * 0
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
		notifier.NotifySuccess(OnMultipliedByZero, r)
	})

	cbus.RegisterUndoDelegationHandler(AddOneUndoDelegation, func(cmd decs.Command, delegate decs.Delegate) {
		r := cmd.Data().(*result)
		fmt.Printf("UNDO %v: %d -> %d\n", cmd.Name(), r.NewValue, r.Value)
		r.NewValue = r.Value
	})

	cbus.RegisterUndoHandler(AddOne, func(cmd decs.Command, delegate decs.Delegate) {
		delegate.To(AddOneUndoDelegation)
	})
	cbus.RegisterUndoHandler(AddTwo, func(cmd decs.Command, delegate decs.Delegate) {
		r := cmd.Data().(*result)
		fmt.Printf("UNDO %v: %d -> %d\n", cmd.Name(), r.NewValue, r.Value)
		r.NewValue = r.Value
	})
	cbus.RegisterUndoHandler(MultiplyByFive, func(cmd decs.Command, delegate decs.Delegate) {
		r := cmd.Data().(*result)
		fmt.Printf("UNDO %v: %d -> %d\n", cmd.Name(), r.NewValue, r.Value)
		r.NewValue = r.Value
	})
	cbus.RegisterUndoHandler(SubtractThree, func(cmd decs.Command, delegate decs.Delegate) {
		r := cmd.Data().(*result)
		fmt.Printf("UNDO %v: %d -> %d\n", cmd.Name(), r.NewValue, r.Value)
		r.NewValue = r.Value
	})

	tests := [][]int{
		{0, 1, 3, 15, 12, 0},
		{-1, 0, 2, 10, 7, 0},
	}

	var r *result
	var testIndex int
	var exp []int
	var i int

	cbus.Subscribe(decs.PurgeApplication, decs.After, func(data interface{}, dispatcher decs.EventDispatcher) {
		r = &result{NewValue: tests[testIndex][0]}
	})

	cbus.Subscribe(OnAddedOne, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnAddedOne, data.(*result).Value, data.(*result).NewValue)
	})
	cbus.Subscribe(OnAddedTwo, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnAddedTwo, data.(*result).Value, data.(*result).NewValue)
	})
	cbus.Subscribe(OnSubtractedThree, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnSubtractedThree, data.(*result).Value, data.(*result).NewValue)
	})
	cbus.Subscribe(OnMultipliedByFive, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnMultipliedByFive, data.(*result).Value, data.(*result).NewValue)
	})
	cbus.Subscribe(OnMultipliedByZero, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnMultipliedByZero, data.(*result).Value, data.(*result).NewValue)
	})

	cbus.Subscribe(OnAddOneUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnAddOneUndone, data.(*result).Value)
	})
	cbus.Subscribe(OnAddTwoUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnAddTwoUndone, data.(*result).Value)
	})
	cbus.Subscribe(OnSubtractThreeUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnSubtractThreeUndone, data.(*result).Value)
	})
	cbus.Subscribe(OnMultiplyByFiveUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnMultiplyByFiveUndone, data.(*result).Value)
	})
	cbus.Subscribe(OnMultiplyByZeroUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnMultiplyByZeroUndone, data.(*result).Value)
	})

	cbus.Subscribe(OnUpdate, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		r = data.(*result)
		go func() {
			defer func() {
				err := recover()
				if err != nil {
					cleanup()
					require.FailNow(t, "%+v", err)
				}
			}()
			wg.Done()
		}()
	})

	for testIndex, exp = range tests {
		waitCnt := 0

		r = &result{Value: exp[0]}

		var cmdDef *decs.CommandDefinition
		for i = 1; i < len(exp); i++ {
			cmdDef = defs[i-1]

			if cbus.CanUndo(cmdDef) {
				waitCnt++
			}

			wg.Add(1)

			// when
			cbus.Do(cbus.NewCommand(cmdDef.Name, r))

			// then
			wg.Wait()
			require.Equal(t, exp[i], r.NewValue)
			r = &result{r.NewValue, r.NewValue}
		}

		for i = len(exp) - 2; i >= 0; i-- {
			cmdDef = defs[i]

			if cbus.CanUndo(cmdDef) {
				wg.Add(1)
			} else {
				wg.Add(waitCnt)
			}

			// when
			cbus.UndoLast()

			// then
			wg.Wait()
			require.Equal(t, exp[i], r.NewValue)
		}

		for i = 1; i < len(exp); i++ {
			cmdDef = defs[i-1]

			wg.Add(1)

			// when
			cbus.RedoLast()

			// then
			wg.Wait()
			require.Equal(t, exp[i], r.NewValue)
		}

		for i = len(exp) - 2; i >= 0; i-- {
			cmdDef = defs[i]

			if cbus.CanUndo(cmdDef) {
				wg.Add(1)
			} else {
				wg.Add(waitCnt)
			}

			// when
			cbus.UndoLast()

			// then
			wg.Wait()
			require.Equal(t, exp[i], r.NewValue)
		}
	}
}

func runTestWithoutUndoHandlers(t *testing.T, cbus *decs.CommandBus) {
	// given
	defs := []*decs.CommandDefinition{
		{
			Name:         AddOne,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnAddOneUndone},
		},
		{
			Name:         AddTwo,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnAddTwoUndone},
		},
		{
			Name:         MultiplyByFive,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnMultiplyByFiveUndone},
		},
		{
			Name:         SubtractThree,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnSubtractThreeUndone},
		},
		{
			Name:         MultiplyByZero,
			DataType:     &result{},
			UndoneEvents: []string{OnUpdate, OnMultiplyByZeroUndone},
		},
	}
	cbus.DefineCommands(defs...)

	cbus.RegisterDelegationHandler(AddOneDelegation1, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value + 1
	})

	cbus.RegisterDelegationHandler(AddOneDelegation2, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		require.Equal(t, r.Value+1, r.NewValue)
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
	})

	cbus.RegisterCommandHandler(AddOne, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		delegate.To(AddOneDelegation1, AddOneDelegation2)
		notifier.NotifySuccess(OnAddedOne, r)
	})
	cbus.RegisterCommandHandler(AddTwo, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value + 2
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
		notifier.NotifySuccess(OnAddedTwo, r)
	})
	cbus.RegisterCommandHandler(MultiplyByFive, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value * 5
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
		notifier.NotifySuccess(OnMultipliedByFive, r)
	})
	cbus.RegisterCommandHandler(SubtractThree, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value - 3
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
		notifier.NotifySuccess(OnSubtractedThree, r)
	})
	cbus.RegisterCommandHandler(MultiplyByZero, func(cmd decs.Command, delegate decs.Delegate, notifier decs.ResultNotifier) {
		r := cmd.Data().(*result)
		r.NewValue = r.Value * 0
		fmt.Printf("DO %v: %d -> %d\n", cmd.Name(), r.Value, r.NewValue)
		notifier.NotifySuccess(OnUpdate, r)
		notifier.NotifySuccess(OnMultipliedByZero, r)
	})

	cbus.RegisterUndoDelegationHandler(AddOneUndoDelegation, func(cmd decs.Command, delegate decs.Delegate) {
		r := cmd.Data().(*result)
		fmt.Printf("UNDO %v: %d -> %d\n", cmd.Name(), r.NewValue, r.Value)
		r.NewValue = r.Value
	})

	tests := [][]int{
		{0, 1, 3, 15, 12, 0},
		{-1, 0, 2, 10, 7, 0},
	}

	var r *result
	var testIndex int
	var exp []int
	var i int

	cbus.Subscribe(decs.PurgeApplication, decs.After, func(data interface{}, dispatcher decs.EventDispatcher) {
		r = &result{NewValue: tests[testIndex][0]}
	})

	cbus.Subscribe(OnAddedOne, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnAddedOne, data.(*result).Value, data.(*result).NewValue)
	})
	cbus.Subscribe(OnAddedTwo, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnAddedTwo, data.(*result).Value, data.(*result).NewValue)
	})
	cbus.Subscribe(OnSubtractedThree, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnSubtractedThree, data.(*result).Value, data.(*result).NewValue)
	})
	cbus.Subscribe(OnMultipliedByFive, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnMultipliedByFive, data.(*result).Value, data.(*result).NewValue)
	})
	cbus.Subscribe(OnMultipliedByZero, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d -> %d\n", OnMultipliedByZero, data.(*result).Value, data.(*result).NewValue)
	})

	cbus.Subscribe(OnAddOneUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnAddOneUndone, data.(*result).Value)
		r.NewValue = r.Value
	})
	cbus.Subscribe(OnAddTwoUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnAddTwoUndone, data.(*result).Value)
		r.NewValue = r.Value
	})
	cbus.Subscribe(OnSubtractThreeUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnSubtractThreeUndone, data.(*result).Value)
		r.NewValue = r.Value
	})
	cbus.Subscribe(OnMultiplyByFiveUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnMultiplyByFiveUndone, data.(*result).Value)
		r.NewValue = r.Value
	})
	cbus.Subscribe(OnMultiplyByZeroUndone, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		fmt.Printf("%v: %d\n", OnMultiplyByZeroUndone, data.(*result).Value)
		r.NewValue = r.Value
	})

	cbus.Subscribe(OnUpdate, decs.AfterSuccess, func(data interface{}, dispatcher decs.EventDispatcher) {
		r = data.(*result)
		go func() {
			defer func() {
				err := recover()
				if err != nil {
					cleanup()
					require.FailNow(t, "%+v", err)
				}
			}()
			wg.Done()
		}()
	})

	for testIndex, exp = range tests {
		r = &result{Value: exp[0]}

		var cmdDef *decs.CommandDefinition
		for i = 1; i < len(exp); i++ {
			cmdDef = defs[i-1]

			wg.Add(1)

			// when
			cbus.Do(cbus.NewCommand(cmdDef.Name, r))

			// then
			wg.Wait()
			require.Equal(t, exp[i], r.NewValue)
			r = &result{r.NewValue, r.NewValue}
		}

		for i = len(exp) - 2; i >= 0; i-- {
			cmdDef = defs[i]

			wg.Add(i + 1)

			// when
			cbus.UndoLast()

			// then
			wg.Wait()
			require.Equal(t, exp[i], r.NewValue)
		}

		for i = 1; i < len(exp); i++ {
			cmdDef = defs[i-1]

			wg.Add(1)

			// when
			cbus.RedoLast()

			// then
			wg.Wait()
			require.Equal(t, exp[i], r.NewValue)
		}

		for i = len(exp) - 2; i >= 0; i-- {
			cmdDef = defs[i]

			wg.Add(i + 1)

			// when
			cbus.UndoLast()

			// then
			wg.Wait()
			require.Equal(t, exp[i], r.NewValue)
		}
	}
}
