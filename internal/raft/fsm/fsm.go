package raft

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
	"github.com/masterkusok/raft-cache/internal/command"
	"github.com/masterkusok/raft-cache/internal/store"
)

type FSM struct {
	storage store.Storage
}

func New(storage store.Storage) *FSM {
	return &FSM{
		storage: storage,
	}
}

func (f *FSM) applyCommand(cmd command.Command) error {
	switch cmd.Action {
	case command.SetAction:
		if err := f.storage.Set(cmd.Key, cmd.Value); err != nil {
			return fmt.Errorf("set value: %w", err)
		}
	case command.DeleteAction:
		if err := f.storage.Delete(cmd.Key); err != nil {
			return fmt.Errorf("delete key: %w", err)
		}
	}
	return nil
}

// func (f *FSM) applyBatch(cmds ...command.Command) error {
// 	errors := make([]error, 0, len(cmds))
// 	for _, cmd := range cmds {
// 		errors = append(errors, f.applyCommand(cmd))
// 	}
// 	return multierr.Combine(errors...)
// }

func (f *FSM) Apply(log *raft.Log) interface{} {
	var cmd command.Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		fmt.Printf("unmarshal command: %v", err)
	}

	if err := f.applyCommand(cmd); err != nil {
		fmt.Printf("execute command: %v", err)
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{
		Data: f.storage.GetSnapshot(),
	}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read snapshot data: %w", err)
	}

	var snapshot snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("unmarshall snapshot data")
	}
	f.storage.ApplySnapshot(snapshot.Data)
	return nil
}
