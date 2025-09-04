package raft

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/raft"
)

type snapshot struct {
	Data map[string]string `json:"data"`
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	var err error
	defer func() {
		if err != nil {
			sink.Cancel()
		}
	}()

	data, err := json.Marshal(s.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	if _, err = sink.Write(data); err != nil {
		return fmt.Errorf("failed to write data to sink: %w", err)
	}

	if err = sink.Close(); err != nil {
		return fmt.Errorf("failed to close sink: %w", err)
	}
	return nil
}

func (s *snapshot) Release() {
}
