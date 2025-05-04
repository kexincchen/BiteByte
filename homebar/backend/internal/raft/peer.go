package raft

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	grpc "github.com/gorilla/rpc/v2"
	jrpc "github.com/gorilla/rpc/v2/json"
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
func NewRaftClient(nodeID string, endpoint string) (*RaftClient, error) {

	return &RaftClient{
		nodeID:     nodeID,
		httpClient: &http.Client{Timeout: RPCTimeout},
		endpoint:   endpoint,
	}, nil
}

// RequestVote sends a RequestVote RPC to a peer
func (c *RaftClient) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	body, err := jrpc.EncodeClientRequest("RaftService.RequestVote", args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return jrpc.DecodeClientResponse(resp.Body, reply)
}

// AppendEntries sends an AppendEntries RPC to a peer
func (c *RaftClient) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	body, err := jrpc.EncodeClientRequest("RaftService.AppendEntries", args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return jrpc.DecodeClientResponse(resp.Body, reply)
}

// RaftService exposes Raft RPCs via HTTP
type RaftService struct {
	node *RaftNode
}

// RegisterRaftService registers the Raft service with an RPC server
func RegisterRaftService(node *RaftNode, rpcServer *grpc.Server) {
	rpcServer.RegisterService(&RaftService{node: node}, "")
}

// SetupRaftRPCServer creates and configures an RPC server for Raft communication
func SetupRaftRPCServer(node *RaftNode) *http.Server {
	rpcServer := grpc.NewServer()
	rpcServer.RegisterCodec(jrpc.NewCodec(), "application/json")
	RegisterRaftService(node, rpcServer)
	mux := http.NewServeMux()
	mux.Handle("/raft", rpcServer)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":808%s", node.id),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	return httpServer
}

func (s *RaftService) RequestVote(r *http.Request, args *RequestVoteArgs, reply *RequestVoteReply) error {
	return s.node.RequestVote(*args, reply)
}

func (s *RaftService) AppendEntries(r *http.Request, args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	return s.node.AppendEntries(*args, reply)
}
