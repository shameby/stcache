package cache

import (
	"io"
	"strings"

	"github.com/hashicorp/raft"
)

type FSM struct {

}

func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	logSlice := strings.Split(string(logEntry.Data), ":")
	if len(logSlice) != 2 {
		return nil
	}
	GetSlot().Set(logSlice[0], logSlice[1])
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{}, nil
}

func (f *FSM) Restore(closer io.ReadCloser) error {
	return GetSlot().FromIO(closer)
}

type snapshot struct {

}

func (ss *snapshot) Persist(sink raft.SnapshotSink) error {
	snapshotBytes, err := GetSlot().ToByte()
	if err != nil {
		_ = sink.Cancel()
		return err
	}
	if _, err := sink.Write(snapshotBytes); err != nil {
		_ = sink.Cancel()
		return err
	}
	if err := sink.Close(); err != nil {
		_ = sink.Cancel()
		return err
	}
	return nil
}

func (ss *snapshot) Release() {

}
