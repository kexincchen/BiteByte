package raft

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
)

// RaftPeer represents a connection to another Raft node
type RaftPeer struct {
	id     string
	client *RaftClient
}

// RaftClient is used to send RPCs to other nodes
type RaftClient struct {
	nodeID     string
	httpClient *http.Client
	endpoint   string
}

// NewRaftClient creates a new client for communicating with a peer node
func NewRaftClient(nodeID string) (*RaftClient, error) {
	// In a real system, we would look up the endpoint from a service registry
	// For now, use a simple mapping (would be replaced with actual node addresses)
	endpoint := fmt.Sprintf("http://localhost:808%s/raft", nodeID)

	return &RaftClient{
		nodeID:     nodeID,
		httpClient: &http.Client{Timeout: RPCTimeout},
		endpoint:   endpoint,
	}, nil
}

// RequestVote sends a RequestVote RPC to a peer
func (c *RaftClient) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Raft-Method", "RequestVote")

	// In a real implementation, encode args to JSON and send in request body
	// For brevity, we'll assume these details are handled

	return nil // Placeholder
}

// AppendEntries sends an AppendEntries RPC to a peer
func (c *RaftClient) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Raft-Method", "AppendEntries")

	// In a real implementation, encode args to JSON and send in request body
	// For brevity, we'll assume these details are handled

	return nil // Placeholder
}

// RaftService exposes Raft RPCs via HTTP
type RaftService struct {
	node *RaftNode
}

// RegisterRaftService registers the Raft service with an RPC server
func RegisterRaftService(node *RaftNode, rpcServer *rpc.Server) {
	rpcServer.RegisterService(&RaftService{node: node}, "")
}

// SetupRaftRPCServer creates and configures an RPC server for Raft communication
func SetupRaftRPCServer(node *RaftNode) *http.Server {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	RegisterRaftService(node, rpcServer)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":808%s", node.id),
		Handler:      rpcServer,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return httpServer
}
