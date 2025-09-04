package raft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/raft"
	"github.com/masterkusok/raft-cache/internal/command"
	fsm "github.com/masterkusok/raft-cache/internal/raft/fsm"
	"github.com/masterkusok/raft-cache/internal/store"
)

const (
	raftTimeout = time.Second * 3
)

type Node struct {
	fsm     *fsm.FSM
	storage store.Storage
	raft    *raft.Raft

	nodeID       string
	leaderAPIUrl string
}

func NewNode(storage store.Storage, fsm *fsm.FSM, leaderUrl, nodeID string) *Node {
	return &Node{
		storage:      storage,
		fsm:          fsm,
		leaderAPIUrl: leaderUrl,
		nodeID:       nodeID,
	}
}

func (n *Node) Open(config Config) error {
	cfg := raft.DefaultConfig()
	cfg.LocalID = raft.ServerID(config.LocalID)

	addr, err := net.ResolveTCPAddr("tcp", config.RaftAddr)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransport(config.RaftAddr, addr, config.MaxPool, config.Timeout, os.Stderr)
	if err != nil {
		return fmt.Errorf("open tcp transport: %w", err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(config.RaftDir, config.SnapshotRetainCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	r, err := raft.NewRaft(cfg, n.fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return fmt.Errorf("create raft: %w", err)
	}

	n.raft = r

	future := r.GetConfiguration()
	if future.Error() != nil {
		return fmt.Errorf("get cluster configuration: %w", err)
	}

	if len(future.Configuration().Servers) != 0 {
		return nil
	}

	if err := n.bootstrapCluster(config.LocalID, transport.LocalAddr()); err != nil {
		log.Fatal("Bootstrap failed:", err)
	}
	return nil
}

func (n *Node) bootstrapCluster(nodeID string, addr raft.ServerAddress) error {
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(nodeID),
				Address: addr,
			},
		},
	}
	return n.raft.BootstrapCluster(configuration).Error()
}

func (n *Node) Join(nodeID, addr string) error {
	configFuture := n.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return fmt.Errorf("get raft configuration: %w", err)
	}

	for _, server := range configFuture.Configuration().Servers {
		if server.ID == raft.ServerID(nodeID) || server.Address == raft.ServerAddress(addr) {
			if server.Address == raft.ServerAddress(addr) && server.ID == raft.ServerID(nodeID) {
				return nil
			}

			future := n.raft.RemoveServer(server.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("remove existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := n.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	return nil
}

func (n *Node) Set(key, value string) error {
	cmd := command.NewSetCommand(key, value)
	if err := n.applyCommand(cmd); err != nil {
		return fmt.Errorf("apply command: %w", err)
	}
	return nil
}

func (n *Node) Delete(key string) error {
	cmd := command.NewDeleteCommand(key)
	if err := n.applyCommand(cmd); err != nil {
		return fmt.Errorf("apply command: %w", err)
	}
	return nil
}

func (n *Node) Get(key string) (string, error) {
	value, err := n.storage.Get(key)
	if err != nil {
		return "", fmt.Errorf("get key '%s': %w", key, err)
	}

	return value, nil
}

func (n *Node) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

func (n *Node) applyCommand(cmd command.Command) error {
	marshaled, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}

	future := n.raft.Apply(marshaled, raftTimeout)
	if err = future.Error(); err != nil {
		return fmt.Errorf("call apply: %w", err)
	}

	return nil
}

func (n *Node) Shutdown(ctx context.Context) error {
	if n.IsLeader() {
		if err := n.raft.LeadershipTransfer().Error(); err != nil {
			return fmt.Errorf("transfer leadership: %w", err)
		}
	}

	if err := n.removeNode(ctx); err != nil {
		return fmt.Errorf("remove node: %w", err)
	}

	if err := n.raft.Shutdown().Error(); err != nil {
		return fmt.Errorf("shutdown raft node: %w", err)
	}
	return nil
}

func (n *Node) removeNode(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/node/%s", n.leaderAPIUrl, n.nodeID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send delete request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad response")
	}
	return nil
}

func (n *Node) RemoveNodeFromCluster(nodeID string) error {
	if n.raft.State() != raft.Leader {
		return fmt.Errorf("only leader can remove nodes")
	}

	if err := n.raft.RemoveServer(raft.ServerID(nodeID), 0, 0).Error(); err != nil {
		return fmt.Errorf("remove server: %v", err)
	}
	return nil
}
