package decs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func (cbus *CommandBus) Load(filename string) error {
	// load commands from disk
	bb, err := ioutil.ReadFile(filename)
	if err != nil {
		cbus.log.Sugar().Error("Cannot read from file", "filename", filename, "error", err)
		return err
	}

	jj := strings.Split(strings.TrimSpace(string(bb)), "\n")

	cc := make([]*command, len(jj))
	for i := 0; i < len(cc); i++ {
		cc[i], err = cbus.UnmarshalCommand([]byte(jj[i]))
		if err != nil {
			cbus.log.Sugar().Error("Cannot unmarshal command", "json", jj[i], "error", err)
			return err
		}
	}

	// reset application
	cbus.ebus.Dispatch(NewEvent(PurgeApplication, nil))
	cbus.versionTracker.update(cbus.id, 0)

	// do all events from scratch
	for _, cmd := range cc {
		cbus.doEventually(cmd, true)
	}

	return nil
}

func (cbus *CommandBus) Save(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		cbus.log.Sugar().Error("Cannot save entire commandStore to file", "filename", filename, "error", err)
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	var e error

	cbus.store.apply(func(cmd *command) bool {
		bb, err := json.Marshal(cmd)
		if err != nil {
			cbus.log.Sugar().Error("Cannot marshal command", "command", cmd, "error", err)
			e = err
			return false
		}

		_, err = f.WriteString(fmt.Sprintf("%s\n", string(bb)))
		if err != nil {
			cbus.log.Sugar().Error("Cannot write to file", "filename", filename, "error", err)
			e = err
			return false
		}

		return true
	})

	return e
}
