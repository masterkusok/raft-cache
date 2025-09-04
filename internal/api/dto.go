package api

type JoinRequest struct {
	Addr   string `json:"addr"`
	NodeID string `json:"node_id"`
}
