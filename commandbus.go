package decs

import (
	"github.com/x-cellent/decs/mq"
	"go.uber.org/zap"
	"time"
)

type CommandBus struct {
	idGenerator         IDGenerator
	id                  string
	log                 *zap.Logger
	mqProvider          mq.Provider
	producer            *producerProxy
	Filters             []Filter
	localFilter         *localCommandFilter
	publishingSuspended bool
	cachingSuspended    bool
	defs                map[string]*CommandDefinition
	guards              map[string][]Guard
	globalGuards        []Guard
	queue               *commandQueue
	incomingCount       int32
	versionTracker      *versionTracker
	handlers            map[string]func(Command, Delegate, ResultNotifier)
	undoHandlers        map[string]func(Command, Delegate)
	undoables           []*command
	redoables           []*command
	flows               map[string][]*command
	store               *commandStore
	Recorder            *recorder
	Player              *player
	ebus                *eventBus
	delegate            *delegate
}

// Use decs.InMemory to keep all commands in memory
func NewDefaultCommandBus(memoryLimit ByteSize) *CommandBus {
	return NewCommandBus(nil, memoryLimit)
}

// Use decs.InMemory to keep all commands in memory
func NewCommandBus(idGenerator IDGenerator, memoryLimit ByteSize) *CommandBus {
	return NewDistributedCommandBus(nil, -1, -1, idGenerator, memoryLimit)
}

// Use decs.InMemory to keep all commands in memory
func NewDefaultDistributedCommandBus(mqProvider mq.Provider, retryInterval, flushDelay time.Duration, memoryLimit ByteSize) *CommandBus {
	return NewDistributedCommandBus(mqProvider, retryInterval, flushDelay, nil, memoryLimit)
}

// Use decs.InMemory to keep all commands in memory
func NewDistributedCommandBus(mqProvider mq.Provider, retryInterval, flushDelay time.Duration, idGenerator IDGenerator, memoryLimit ByteSize) *CommandBus {
	if idGenerator == nil {
		idGenerator = NewDefaultIDGenerator()
	}

	ebus := newEventBus()

	cbusID := idGenerator.Create()

	cbus := &CommandBus{
		idGenerator:    idGenerator,
		id:             cbusID,
		log:            zap.NewNop(),
		defs:           make(map[string]*CommandDefinition),
		guards:         make(map[string][]Guard),
		queue:          newCommandQueue(),
		versionTracker: newVersionTracker(cbusID),
		handlers:       make(map[string]func(Command, Delegate, ResultNotifier)),
		undoHandlers:   make(map[string]func(Command, Delegate)),
		flows:          make(map[string][]*command),
		ebus:           ebus,
		delegate:       newDelegate(ebus),
	}

	cbus.SetMqProvider(mqProvider, retryInterval, flushDelay)

	ebus.commandBus = cbus

	cbus.HandleLocalCommandsImmediately()

	cbus.Recorder = newRecorder(cbus.log)
	cbus.Player = newPlayer(cbus)
	cbus.store = NewCommandStore(cbus, memoryLimit)

	cbus.Player.BeforePlayBack = func(cmd Command) {
		cbus.SuspendPublishing()
		cbus.SuspendCaching()
	}
	cbus.Player.AfterPlayBack = func(cmd Command) {
		cbus.ResumeCaching()
		cbus.ResumePublishing()
	}

	go cbus.handleCommands()

	return cbus
}
