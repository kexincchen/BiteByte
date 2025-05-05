package raft

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ClusterState represents the overall state of the Raft cluster
type ClusterState struct {
	LeaderID    string
	Term        uint64
	CommitIndex uint64
	Nodes       map[string]NodeStatus
	LastUpdated time.Time
}

// NodeStatus represents the status of a node in the cluster
type NodeStatus struct {
	ID        string
	State     NodeState
	IsHealthy bool
	LastSeen  time.Time
	Address   string
}

// ClusterCoordinator manages the Raft cluster membership and monitoring
type ClusterCoordinator struct {
	mu         sync.RWMutex
	nodes      map[string]*RaftNode
	state      ClusterState
	httpServer *http.Server
	logger     *log.Logger
	stopCh     chan struct{}
	peerAddrs  map[string]string
	selfID     string
}

// NewClusterCoordinator creates a new coordinator for managing the cluster
func NewClusterCoordinator(logger *log.Logger, peerAddrs map[string]string) *ClusterCoordinator {
	return &ClusterCoordinator{
		nodes: make(map[string]*RaftNode),
		state: ClusterState{
			LeaderID:    "",
			Term:        0,
			CommitIndex: 0,
			Nodes:       make(map[string]NodeStatus),
			LastUpdated: time.Now(),
		},
		logger:    logger,
		stopCh:    make(chan struct{}),
		peerAddrs: peerAddrs,
	}
}

// RegisterNode adds a node to the coordinator's management
func (c *ClusterCoordinator) RegisterNode(node *RaftNode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	selfAddr := "http://127.0.0.1" + port

	c.nodes[node.id] = node
	c.state.Nodes[node.id] = NodeStatus{
		ID:        node.id,
		State:     Follower, // Assume follower initially
		IsHealthy: true,
		LastSeen:  time.Now(),
		Address:   selfAddr,
	}

	for id, raftAddr := range c.peerAddrs {
		if id == node.id {
			continue
		}
		if _, ok := c.state.Nodes[id]; ok {
			continue
		}

		businessAddr := raftToBusinessAddr(raftAddr)
		c.state.Nodes[id] = NodeStatus{
			ID:        id,
			State:     Follower,
			IsHealthy: false,
			LastSeen:  time.Time{},
			Address:   businessAddr,
		}
	}
}

// Start begins the coordinator's monitoring and management tasks
func (c *ClusterCoordinator) Start(ctx context.Context, nodeID string) error {
	c.selfID = nodeID
	// Start HTTP server for admin API
	c.startHTTPServer(nodeID)

	// Start periodic health checks and state updates
	go c.runMonitoring(ctx)

	return nil
}

// Stop gracefully shuts down the coordinator
func (c *ClusterCoordinator) Stop() error {
	close(c.stopCh)

	if c.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return c.httpServer.Shutdown(ctx)
	}

	return nil
}

// GetClusterState returns the current state of the cluster
func (c *ClusterCoordinator) GetClusterState() ClusterState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid concurrent modification
	state := c.state
	state.Nodes = make(map[string]NodeStatus, len(c.state.Nodes))
	for id, status := range c.state.Nodes {
		state.Nodes[id] = status
	}

	return state
}

// updateNodeStatus updates the status of a node in the cluster state
func (c *ClusterCoordinator) updateNodeStatus(nodeID string, status NodeStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state.Nodes[nodeID] = status
	c.state.LastUpdated = time.Now()

	// If this node is leader, update leader ID
	if status.State == Leader {
		c.state.LeaderID = nodeID
	} else if c.state.LeaderID == nodeID {
		// If current leader is no longer leader, clear leader ID
		c.state.LeaderID = ""
	}
}

// runMonitoring periodically checks node health and updates cluster state
func (c *ClusterCoordinator) runMonitoring(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.checkNodesHealth()
			c.updateClusterState()

		case <-ctx.Done():
			return

		case <-c.stopCh:
			return
		}
	}
}

// checkNodesHealth checks the health of all nodes in the cluster
func (c *ClusterCoordinator) checkNodesHealth() {
	c.mu.RLock()
	nodes := make([]*RaftNode, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	c.mu.RUnlock()

	for _, node := range nodes {
		// For local nodes, we can check directly
		// For remote nodes, we would use the RPC client

		// Update node status
		c.mu.Lock()
		status := c.state.Nodes[node.id]
		status.LastSeen = time.Now()
		status.IsHealthy = true // Assume healthy if we can access it

		// Get state from node (for local nodes)
		node.mu.Lock()
		status.State = node.state
		node.mu.Unlock()

		c.state.Nodes[node.id] = status
		c.mu.Unlock()
	}

	c.mu.RLock()
	peers := make(map[string]NodeStatus, len(c.state.Nodes))
	for id, st := range c.state.Nodes {
		peers[id] = st
	}
	c.mu.RUnlock()

	for id, st := range peers {
		if id == c.selfID {
			continue
		}
		url := strings.TrimRight(st.Address, "/") + "/health"

		ok := probe(url)
		c.mu.Lock()
		cur := c.state.Nodes[id]
		cur.IsHealthy = ok
		cur.LastSeen = time.Now()
		c.state.Nodes[id] = cur
		c.mu.Unlock()
	}
}

// updateClusterState collects and updates the overall cluster state
func (c *ClusterCoordinator) updateClusterState() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find the leader and its term
	var maxTerm uint64
	for _, node := range c.nodes {
		node.mu.Lock()

		if node.state == Leader {
			c.state.LeaderID = node.id
			maxTerm = node.currentTerm
			c.state.CommitIndex = node.commitIndex
		} else if node.currentTerm > maxTerm {
			maxTerm = node.currentTerm
		}

		node.mu.Unlock()
	}

	c.state.Term = maxTerm
	c.state.LastUpdated = time.Now()
}

// startHTTPServer starts an HTTP server for administrative API endpoints
func (c *ClusterCoordinator) startHTTPServer(nodeID string) {
	mux := http.NewServeMux()

	// Add endpoints for cluster management
	mux.HandleFunc("/cluster/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		state := c.GetClusterState()

		alive := 0
		for _, ns := range state.Nodes {
			if ns.IsHealthy {
				alive++
			}
		}
		fmt.Fprintf(w, `{"leader":"%s","term":%d,"nodes":%d}`,
			state.LeaderID, state.Term, alive)
	})

	mux.HandleFunc("/cluster/nodes", func(w http.ResponseWriter, r *http.Request) {
		// Return info about all nodes
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		state := c.GetClusterState()
		fmt.Fprintf(w, `{"count":%d}`, len(state.Nodes))
	})

	// Use a different coordinator port for each node
	coordPort := 8090 + int(nodeID[0]-'0') // Assumes nodeID is a single digit

	// Start the HTTP server
	c.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", coordPort),
		Handler: mux,
	}

	go func() {
		if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.logger.Printf("HTTP server error: %v", err)
		}
	}()
}

func probe(url string) bool {
	cli := &http.Client{Timeout: 300 * time.Millisecond}
	resp, err := cli.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func raftToBusinessAddr(raftAddr string) string {
	u, _ := url.Parse(strings.TrimSuffix(raftAddr, "/raft"))
	host, portStr, _ := strings.Cut(u.Host, ":")
	port, _ := strconv.Atoi(portStr)
	businessPort := port + 920
	return fmt.Sprintf("http://%s:%d", host, businessPort)
}
