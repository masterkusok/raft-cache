package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/masterkusok/raft-cache/internal/raft"
)

type Server struct {
	node *raft.Node
}

func NewServer(node *raft.Node) *Server {
	return &Server{
		node: node,
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
	r.Get("/api/v1/key/{key}/{value}", s.handleGet)
	r.Delete("/api/v1/key/{key}/{value}", s.handleDelete)

	http.ListenAndServe(port, r)
}

func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("set request recieved\n")
	key := chi.URLParam(r, "key")
	value := chi.URLParam(r, "value")

	if !s.node.IsLeader() {
		fmt.Printf("node should be leader")
		w.WriteHeader(http.StatusMisdirectedRequest)
		return
	}

	if err := s.node.Set(key, value); err != nil {
		fmt.Printf("failed to set key '%s' value '%s': %v\n", key, value, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("set success\n")
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("get request recieved\n")
	key := chi.URLParam(r, "key")

	var val string
	var err error

	if val, err = s.node.Get(key); err != nil {
		fmt.Printf("failed to get key '%s' value: %v\n", key, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s\n", val)
	fmt.Printf("get success\n")
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("join node\n")
	var request JoinRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = s.node.Join(request.NodeID, request.Addr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("joined successfully")
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {

}
