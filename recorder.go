package decs

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
)

type Record struct {
	Commands []*command `json:"commands"`
}

func NewRecord() *Record {
	return &Record{}
}

func (r *Record) Clear() {
	r.Commands = r.Commands[:0]
}

type recorder struct {
	log       *zap.Logger
	recording bool
	record    *Record
}

func newRecorder(log *zap.Logger) *recorder {
	return &recorder{
		log:       log,
		recording: false,
		record:    NewRecord(),
	}
}

func (r *recorder) HasRecord() bool {
	return len(r.record.Commands) > 0
}

func (r *recorder) Reset() {
	r.record.Clear()
}

func (r *recorder) SetReplay(record *Record) {
	r.record = record
}

func (r *recorder) Start() {
	if !r.IsRecording() {
		r.Reset()
		r.Resume()
	}
}

func (r *recorder) Stop() {
	r.Suspend()
}

func (r *recorder) Suspend() {
	r.recording = false
}

func (r *recorder) Resume() {
	r.recording = true
}

func (r *recorder) IsRecording() bool {
	return r.recording
}

func (r *recorder) append(cmd *command) {
	r.record.Commands = append(r.record.Commands, cmd)
}

func (r *recorder) SaveRecord(filename string) (bool, error) {
	if !r.HasRecord() {
		return false, nil
	}

	bb, err := json.Marshal(r.record)
	if err != nil {
		r.log.Sugar().Error("Failed to marshal record", "record", r.record, "error", err)
		return false, err
	}

	err = ioutil.WriteFile(filename, bb, 0644)
	if err != nil {
		r.log.Sugar().Error("Failed to save record", "filename", filename, "error", err)
		return false, err
	}

	return true, nil
}
