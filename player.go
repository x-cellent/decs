package decs

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

const interval = 500 * time.Millisecond

type player struct {
	cbus           *CommandBus
	replay         *Record
	index          int
	playing        bool
	finished       bool
	speed          time.Duration
	BeforePlayBack func(command Command)
	AfterPlayBack  func(command Command)
}

func newPlayer(cbus *CommandBus) *player {
	return &player{
		cbus:     cbus,
		replay:   NewRecord(),
		index:    0,
		playing:  false,
		finished: true,
		speed:    interval,
	}
}

func (p *player) HasReplay() bool {
	return len(p.replay.Commands) > 0
}

func (p *player) IsPlaying() bool {
	return p.playing
}

func (p *player) HasFinished() bool {
	return p.finished
}

func (p *player) SetSpeed(speed time.Duration) {
	p.speed = speed
}

func (p *player) DecreaseSpeed() {
	p.speed += interval
}

func (p *player) IncreaseSpeed() {
	if p.speed-interval < 0 {
		return
	}
	p.speed -= interval
}

func (p *player) Reset() {
	p.replay.Clear()
}

func (p *player) SetReplay(replay *Record) {
	p.replay = replay

	if p.IsPlaying() {
		p.Restart()
	}
}

func (p *player) LoadRecord(filename string) error {
	bb, err := ioutil.ReadFile(filename)
	if err != nil {
		p.cbus.log.Sugar().Error("Failed to load record", "filename", filename, "error", err)
		return err
	}

	R := NewRecord()
	err = json.Unmarshal(bb, R)
	if err != nil {
		p.cbus.log.Sugar().Error("Failed to unmarshal record", "filename", filename, "error", err)
		return err
	}

	p.SetReplay(R)

	return nil
}

func (p *player) Start() {
	if !p.IsPlaying() {
		p.Restart()
	}
}

func (p *player) Stop() {
	p.Pause()
	p.index = 0
}

func (p *player) Pause() {
	p.playing = false
}

func (p *player) Restart() {
	p.Stop()
	if !p.HasReplay() {
		return
	}

	p.finished = false
	p.playing = true

	go p.run()
}

func (p *player) run() {
	done := make(chan bool, 1)
	for !p.finished {
		select {
		case <-done:
			p.finished = true
			p.playing = false
		case <-time.After(p.speed):
			if p.IsPlaying() {
				if p.index < 0 || p.index >= len(p.replay.Commands) {
					done <- true
				} else {
					cmd := p.replay.Commands[p.index]
					if p.BeforePlayBack != nil {
						p.BeforePlayBack(cmd)
					}
					p.cbus.do(cmd)
					if p.AfterPlayBack != nil {
						p.AfterPlayBack(cmd)
					}
					p.index++
				}
			}
		}
	}
}
