package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/masterkusok/raft-cache/internal/api"
	"github.com/masterkusok/raft-cache/internal/raft"
	fsm "github.com/masterkusok/raft-cache/internal/raft/fsm"
	"github.com/masterkusok/raft-cache/internal/store"
)

func getRaftraftCfg() raft.Config {
	localID := flag.String("local-id", "node-1", "Local node ID")
	raftAddr := flag.String("raft-addr", "localhost:8081", "Raft server address")
	raftDir := flag.String("raft-dir", "temp/", "Raft data directory")
	leaderAddr := flag.String("leader-addr", "", "Leader address for joining cluster")
	leaderEndpoint := flag.String("leader-api-endpoint", "", "Leader api endpoint for joining cluster")
	maxPool := flag.Int("max-pool", 3, "Maximum connection pool size")
	timeout := flag.Duration("timeout", 3*time.Second, "Raft timeout duration")
	snapshotRetain := flag.Int("snapshot-retain", 2, "Number of snapshots to retain")
	port := flag.String("port", ":8000", "api port")

	flag.Parse()

	return raft.Config{
		LocalID:             *localID,
		RaftAddr:            *raftAddr,
		RaftDir:             *raftDir,
		LeaderAddr:          *leaderAddr,
		LeaderApiEndpoint:   *leaderEndpoint,
		MaxPool:             *maxPool,
		Timeout:             *timeout,
		SnapshotRetainCount: *snapshotRetain,
		Port:                *port,
	}
}

func main() {
	raftCfg := getRaftraftCfg()

	store := store.NewInMemoryStorage()
	fsm := fsm.New(store)

	node := raft.NewNode(store, fsm, raftCfg.LeaderApiEndpoint, raftCfg.LocalID)

	if err := node.Open(raftCfg); err != nil {
		log.Fatalf("failed to start node: %v", err)
	}

	if raftCfg.LeaderAddr != "" {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := join(ctx, raftCfg.LeaderApiEndpoint, raftCfg.RaftAddr, raftCfg.LocalID); err != nil {
			panic(err)
		}
	}

	server := api.NewServer(node)
	server.Start(raftCfg.Port)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := cleanup(ctx, node); err != nil {
		fmt.Printf("failed to cleanup node: %v", err)
		return
	}

	fmt.Printf("node stopped\n")

}

func join(ctx context.Context, endpoint, addr, nodeID string) error {
	request := api.JoinRequest{
		Addr:   addr,
		NodeID: nodeID,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal join request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/node", endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do http request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("http request with unsuccessful code")
	}

	return nil
}

func cleanup(ctx context.Context, node *raft.Node) error {
	return node.Shutdown(ctx)
}
