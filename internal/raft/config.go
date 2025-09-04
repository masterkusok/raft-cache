package raft

import "time"

type Config struct {
	LocalID             string        `yaml:"local-id"`
	RaftAddr            string        `yaml:"raft-addr"`
	RaftDir             string        `yaml:"raft-dir"`
	LeaderAddr          string        `yaml:"leader-addr,omitempty"`
	LeaderApiEndpoint   string        `yaml:"leader-api-endpoint"`
	MaxPool             int           `yaml:"max-pool"`
	Timeout             time.Duration `yaml:"timeout"`
	SnapshotRetainCount int           `yaml:"snapshot-retain-count"`
	Port                string        `yaml:"port"`
}
