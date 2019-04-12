package decs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type ByteSize uint64

const (
	Byte ByteSize = 1 << (10 * iota)
	KiloByte
	MegaByte
	GigaByte
	TeraByte
	PetaByte
	ExaByte
)

const (
	CommandStoreFile              = ".commandStore"
	CommandStoreFilePerm          = 0644
	InMemory             ByteSize = 0
)

type commandStore struct {
	cbus        *CommandBus
	commands    []*command
	size        int
	memoryLimit ByteSize
	swap        string
}

func (b ByteSize) NumBytes() int {
	return int(b)
}

func (b ByteSize) String() string {
	switch {
	case b >= ExaByte:
		return fmt.Sprintf("%.2fEB", float64(b)/float64(ExaByte))
	case b >= PetaByte:
		return fmt.Sprintf("%.2fPB", float64(b)/float64(PetaByte))
	case b >= TeraByte:
		return fmt.Sprintf("%.2fTB", float64(b)/float64(TeraByte))
	case b >= GigaByte:
		return fmt.Sprintf("%.2fGB", float64(b)/float64(GigaByte))
	case b >= MegaByte:
		return fmt.Sprintf("%.2fMB", float64(b)/float64(MegaByte))
	case b >= KiloByte:
		return fmt.Sprintf("%.2fKB", float64(b)/float64(KiloByte))
	}
	return fmt.Sprintf("%dB", b)
}

func NewCommandStore(cbus *CommandBus, memoryLimit ByteSize) *commandStore {
	return &commandStore{
		cbus:        cbus,
		memoryLimit: memoryLimit,
		swap:        CommandStoreFile,
	}
}

func (cs *commandStore) MemoryLimit() ByteSize {
	return cs.memoryLimit
}

func (cs *commandStore) LimitMemory(limit ByteSize) {
	cs.memoryLimit = limit
}

func (cs *commandStore) Swap() string {
	return cs.swap
}

func (cs *commandStore) moveSwap() {
	src := cs.swap

	_, err := os.Stat(src)
	if err != nil {
		return
	}

	from, err := os.Open(src)
	if err != nil {
		cs.cbus.log.Sugar().Error("Cannot open file", "filename", src, "error", err)
		return
	}
	defer func() {
		_ = from.Close()
		_ = os.Remove(src)
	}()

	dest := src + ".bak"

	to, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, CommandStoreFilePerm)
	if err != nil {
		cs.cbus.log.Sugar().Error("Cannot open file", "filename", dest, "error", err)
		return
	}
	defer func() {
		_ = to.Close()
	}()

	_, err = io.Copy(to, from)
	if err != nil {
		cs.cbus.log.Sugar().Error("Cannot copy file", "source", src, "destination", dest, "error", err)
		return
	}

	cs.swap = dest
}

func (cs *commandStore) Apply(f func(cmd Command) bool) {
	cs.apply(func(cmd *command) bool {
		return f(cmd)
	})
}

func (cs *commandStore) ApplyExclusively(f func(cmd Command) bool) {
	cs.applyExclusively(func(cmd *command) bool {
		return f(cmd)
	})
}

func (cs *commandStore) apply(f func(cmd *command) bool) {
	cs._apply(f)
}

func (cs *commandStore) applyExclusively(f func(cmd *command) bool) {
	cs._apply(f)
}

func (cs *commandStore) _apply(f func(cmd *command) bool) {
	file, err := os.Open(cs.swap)
	if err == nil {
		defer func() {
			_ = file.Close()
		}()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			cmd, err := cs.cbus.UnmarshalCommand(scanner.Bytes())
			if err != nil {
				cs.cbus.log.Sugar().Error("Cannot unmarshal command", "json", scanner.Text(), "error", err)
				return
			}

			if !f(cmd) {
				return
			}
		}
	}

	for _, cmd := range cs.commands {
		if !f(cmd) {
			return
		}
	}
}

func (cs *commandStore) Clear() {
	cs._clear()
}

func (cs *commandStore) _clear() {
	_ = os.Remove(cs.swap)
	cs.commands = cs.commands[:0]
	cs.size = 0
}

func (cs *commandStore) push(cmd *command) {
	var i int

	for i = len(cs.commands); i > 0; i-- {
		if cs.commands[i-1].Version_ < cmd.Version_ {
			break
		}
	}

	// this often occurring special case increases performance
	if i == len(cs.commands) {
		cs.commands = append(cs.commands, cmd)
	} else {
		cs.commands = append(cs.commands[:i], append([]*command{cmd}, cs.commands[i:]...)...)
	}

	cs.size += cmd.numBytes

	if cs.memoryLimit > 0 && cs.ExceedsMemoryLimit() {
		cs._writeHalfToDisk()
	}
}

func (cs *commandStore) delete(cmd *command) *command {
	var c *command

	if cs.memoryLimit > 0 && len(cs.commands) == 0 {
		cs._populateHalfFromDisk()
	}

	for i := len(cs.commands) - 1; i >= 0; i-- {
		if cs.commands[i].Id == cmd.Id {
			c = cs.commands[i]
			cs.size -= c.numBytes
			cs.commands = append(cs.commands[:i], cs.commands[i+1:]...)
			break
		}
	}

	return c
}

func (cs *commandStore) ExceedsMemoryLimit() bool {
	return cs.size > cs.memoryLimit.NumBytes()
}

func (cs *commandStore) halfSize() ByteSize {
	return cs.memoryLimit/2 + cs.memoryLimit%2
}

func (cs *commandStore) _writeHalfToDisk() {
	halfSize := cs.halfSize().NumBytes()
	halfIdx := 0
	sum := 0
	for i, c := range cs.commands {
		if sum >= halfSize {
			break
		}
		halfIdx = i + 1
		sum += c.numBytes
	}

	jj := make([]string, halfIdx)
	for i := 0; i < halfIdx; i++ {
		bb, err := json.Marshal(cs.commands[i])
		if err != nil {
			cs.cbus.log.Sugar().Error("Cannot marshal command", "command", cs.commands[i], "error", err)
			return
		}
		jj[i] = string(bb)
		cs.size -= cs.commands[i].numBytes
	}

	f, err := os.OpenFile(cs.swap, os.O_CREATE|os.O_APPEND|os.O_WRONLY, CommandStoreFilePerm)
	if err != nil {
		cs.cbus.log.Sugar().Error("Cannot open commandStore file", "error", err)
		return
	}

	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(fmt.Sprintf("%s\n", strings.Join(jj, "\n")))
	if err != nil {
		cs.cbus.log.Sugar().Error("Cannot append to commandStore file", "error", err)
		return
	}

	cs.commands = cs.commands[halfIdx:]
}

func (cs *commandStore) _populateHalfFromDisk() {
	bb, err := ioutil.ReadFile(cs.swap)
	if err != nil {
		cs.cbus.log.Sugar().Error("Cannot read from commandStore file", "error", err)
		return
	}

	jj := strings.Split(strings.TrimSpace(string(bb)), "\n")
	size := 0
	for _, c := range jj {
		size += len(c)
	}

	halfSize := cs.halfSize().NumBytes()
	if halfSize > size {
		halfSize = size
	}

	halfIdx := len(jj)
	sum := size
	for i := len(jj) - 1; i >= 0; i-- {
		if sum < halfSize {
			break
		}
		halfIdx = i + 1
		sum -= len(jj[i])
	}

	from := len(jj) - halfIdx

	cs.commands = make([]*command, halfIdx)
	cs.size = 0
	for i := 0; i < halfIdx; i++ {
		cs.commands[i], err = cs.cbus.UnmarshalCommand([]byte(jj[from+i]))
		if err != nil {
			cs.cbus.log.Sugar().Error("Cannot unmarshal command", "json", jj[from+i], "error", err)
			return
		}
		cs.size += cs.commands[i].numBytes
	}

	f, err := os.OpenFile(cs.swap, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, CommandStoreFilePerm)
	if err != nil {
		cs.cbus.log.Sugar().Error("Cannot open commandStore file", "error", err)
		return
	}

	defer func() {
		_ = f.Close()
	}()

	if from == 0 {
		return
	}

	_, err = f.WriteString(strings.Join(jj[:from], "\n"))
	if err != nil {
		cs.cbus.log.Sugar().Error("Cannot write to commandStore file", "error", err)
		return
	}
}
