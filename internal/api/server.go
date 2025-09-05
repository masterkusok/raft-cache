package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/masterkusok/raft-cache/internal/raft"
	"github.com/sirupsen/logrus"
)

type Server struct {
	node   *raft.Node
	logger *logrus.Logger
}

func NewServer(node *raft.Node, logger *logrus.Logger) *Server {
	return &Server{
		node:   node,
		logger: logger,
	}
}

func (s *Server) Start(port string) {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.Logger,
		middleware.RequestID,
	)

	r.Post("/api/v1/key/{key}/{value}", s.handleSet)
	r.Post("/api/v1/node", s.handleJoin)

	r.Get("/api/v1/key/{key}", s.handleGet)

	r.Delete("/api/v1/key/{key}", s.handleDelete)
	r.Delete("/api/v1/node/{id}", s.handleRemoveNode)

	go func() {
		s.logger.Info("starting http server...")
		if err := http.ListenAndServe(port, r); err != nil {
			s.logger.Fatalf("HTTP server: %v", err)
		}
	}()
}

func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		s.logger.Warnf("no key in set request provided")
		renderAPIError(w, http.StatusBadRequest, "key is required")
		return
	}

	value := chi.URLParam(r, "value")
	if value == "" {
		s.logger.Warnf("no value in set request provided")
		renderAPIError(w, http.StatusBadRequest, "value is required")
		return
	}

	if !s.node.IsLeader() {
		s.logger.Warnf("set request to follower node")
		renderAPIError(w, http.StatusMisdirectedRequest, "set requests to follower node are not allowed")
		return
	}

	if err := s.node.Set(key, value); err != nil {
		s.logger.Errorf("failed to set value: %v", err)
		renderAPIError(w, http.StatusMisdirectedRequest, "cannot set value")
		return
	}
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		s.logger.Warnf("no key in get request provided")
		renderAPIError(w, http.StatusBadRequest, "key is required")
		return
	}

	var val string
	var err error

	if val, err = s.node.Get(key); err != nil {
		s.logger.Warnf("failed to get value from store: %v", err)
		renderAPIError(w, http.StatusInternalServerError, "cannot get value from store")
		return
	}

	fmt.Fprintf(w, "%s\n", val)
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	var request JoinRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Errorf("failed to read request body: %v", err)
		renderAPIError(w, http.StatusInternalServerError, "failed to parse request")
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		s.logger.Errorf("failed to unmarshal request: %v", err)
		renderAPIError(w, http.StatusInternalServerError, "failed to parse request")
		return
	}

	s.logger.Infof("joining node with id '%s'", request.NodeID)
	if err = s.node.Join(request.NodeID, request.Addr); err != nil {
		s.logger.Errorf("failed join to cluster: %v", err)
		renderAPIError(w, http.StatusInternalServerError, "connection to cluster failed")
		return
	}
	s.logger.Infof("node with id '%s' joined successfully", request.NodeID)
}

func (s *Server) handleRemoveNode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		s.logger.Warnf("id is required to remove node")
		renderAPIError(w, http.StatusBadRequest, "id is required to remove node")
		return
	}

	s.logger.Infof("removing node with id '%s'", id)
	if err := s.node.RemoveNodeFromCluster(id); err != nil {
		s.logger.Errorf("failed to remove node from cluster: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.logger.Infof("successfully removed node with id '%s'", id)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		s.logger.Warnf("failed to get key from url")
		renderAPIError(w, http.StatusBadRequest, "key is required")
		return
	}

	if err := s.node.Delete(key); err != nil {
		s.logger.Errorf("failed to delete key: %v", err)
		renderAPIError(w, http.StatusInternalServerError, fmt.Sprintf("cannot delete key: %s", key))
		return
	}
}
