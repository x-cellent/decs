package decs

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func (cbus *CommandBus) UnmarshalCommand(bb []byte) (*command, error) {
	cmd := &command{bus: cbus}
	err := cmd.UnmarshalJSON(bb)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (c *command) UnmarshalJSON(bb []byte) error {
	c.numBytes = len(bb)

	raw := make(map[string]*json.RawMessage)

	err := json.Unmarshal(bb, &raw)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["id"], &c.Id)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["version"], &c.Version_)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["aggID"], &c.AggregateId)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["name"], &c.Name_)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["category"], &c.Cat)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["cbusID"], &c.BusId)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["raised"], &c.Raised_)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["undo"], &c.Undo)
	if err != nil {
		return err
	}

	err = json.Unmarshal(*raw["oneshot"], &c.Oneshot)
	if err != nil {
		return err
	}

	if c.bus == nil {
		return errors.New("command has no reference to command bus")
	}
	def := c.bus.defs[c.Name_]
	if def == nil {
		return fmt.Errorf("command %q is not registered yet and therefore cannot be unmarshalled", c.Name_)
	}
	typ := reflect.TypeOf(def.DataType)
	if typ == nil {
		return nil
	}
	if strings.HasPrefix(typ.String(), "*") {
		typ = typ.Elem()
	}

	data := raw["data"]
	if data != nil {
		c.Data_ = reflect.New(typ).Elem().Addr().Interface()
		err := json.Unmarshal(*data, &c.Data_)
		if err != nil {
			return err
		}
	}

	return nil
}
