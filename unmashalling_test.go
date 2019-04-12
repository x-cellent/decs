package decs

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEventMarshalling(t *testing.T) {
	// given
	cbus := NewDefaultCommandBus(InMemory)
	cbus.DefineCommands(&CommandDefinition{
		Name:     test,
		DataType: &userData{},
	})

	cmd := cbus.NewCommand(test, &userData{
		Name:   "Buddy",
		Human:  true,
		Ignore: "ignore",
		ignore: -1,
	})

	// when
	JSON, err := json.Marshal(cmd)

	// then
	require.Nil(t, err)

	// when
	c, err := cbus.UnmarshalCommand(JSON)

	// then
	require.Nil(t, err)
	require.NotNil(t, c)

	require.NotNil(t, c.Data())
	require.Equal(t, "&{Buddy true  0}", fmt.Sprintf("%v", c.Data()))
	d, ok := c.Data().(*userData)
	require.True(t, ok)
	require.Equal(t, "Buddy", d.Name)
	require.Equal(t, true, d.Human)
	require.Equal(t, "", d.Ignore)
	require.Equal(t, 0, d.ignore)
}

func TestRecordMarshalling(t *testing.T) {
	// given
	cbus := NewDefaultCommandBus(InMemory)
	cbus.DefineCommand(&CommandDefinition{
		Name:     test,
		DataType: &userData{},
	})

	cmd1 := cbus.NewCommand(test, &userData{
		Name:   "Bud",
		Human:  true,
		Ignore: "Spencer",
		ignore: -1,
	})
	cmd2 := cbus.NewCommand(test, &userData{
		Name:   "Terence",
		Human:  true,
		Ignore: "Hill",
		ignore: -1,
	})
	record := &Record{Commands: []*command{cmd1, cmd2}}

	// when
	JSON, err := json.Marshal(record)

	// then
	require.Nil(t, err)

	// when
	record.Clear()
	//record = NewRecord()
	err = json.Unmarshal(JSON, record)

	// then
	require.Nil(t, err)
	require.NotNil(t, record)
	require.Equal(t, 2, len(record.Commands))

	c := record.Commands[0]

	require.NotNil(t, c.Data())
	require.Equal(t, "&{Bud true  0}", fmt.Sprintf("%v", c.Data()))
	d, ok := c.Data().(*userData)
	require.True(t, ok)
	require.Equal(t, "Bud", d.Name)
	require.Equal(t, true, d.Human)
	require.Equal(t, "", d.Ignore)
	require.Equal(t, 0, d.ignore)

	c = record.Commands[1]

	require.NotNil(t, c.Data())
	require.Equal(t, "&{Terence true  0}", fmt.Sprintf("%v", c.Data()))
	d, ok = c.Data().(*userData)
	require.True(t, ok)
	require.Equal(t, "Terence", d.Name)
	require.Equal(t, true, d.Human)
	require.Equal(t, "", d.Ignore)
	require.Equal(t, 0, d.ignore)
}

type userData struct {
	Name   string `json:"name"`
	Human  bool   `json:"human"`
	Ignore string `json:"-"`
	ignore int
}
