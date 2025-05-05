package raft

import (
	"context"
	"fmt"
	// "log"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"
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
	logger     *zerolog.Logger
	stopCh     chan struct{}
}

// NewClusterCoordinator creates a new coordinator for managing the cluster
func NewClusterCoordinator(logger *zerolog.Logger) *ClusterCoordinator {
	return &ClusterCoordinator{
		nodes: make(map[string]*RaftNode),
		state: ClusterState{
			LeaderID:    "",
			Term:        0,
			CommitIndex: 0,
			Nodes:       make(map[string]NodeStatus),
			LastUpdated: time.Now(),
		},
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// RegisterNode adds a node to the coordinator's management
func (c *ClusterCoordinator) RegisterNode(node *RaftNode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodes[node.id] = node
	c.state.Nodes[node.id] = NodeStatus{
		ID:        node.id,
		State:     Follower, // Assume follower initially
		IsHealthy: true,
		LastSeen:  time.Now(),
		Address:   fmt.Sprintf("localhost:808%s", node.id), // Simplified address for demo
	}
}

// Start begins the coordinator's monitoring and management tasks
func (c *ClusterCoordinator) Start(ctx context.Context, nodeID string) error {
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
}

// updateClusterState collects and updates the overall cluster state
func (c *ClusterCoordinator) updateClusterState() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find the leader and its term
	for _, node := range c.nodes {
		node.mu.Lock()

		if node.state == Leader {
			c.state.LeaderID = node.id
			c.state.Term = node.currentTerm
			c.state.CommitIndex = node.commitIndex
		}

		node.mu.Unlock()
	}

	c.state.LastUpdated = time.Now()
}

// startHTTPServer starts an HTTP server for administrative API endpoints
func (c *ClusterCoordinator) startHTTPServer(nodeID string) {
	mux := http.NewServeMux()

	// Add endpoints for cluster management
	mux.HandleFunc("/cluster/status", func(w http.ResponseWriter, r *http.Request) {
		// Return cluster status as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		state := c.GetClusterState()
		fmt.Fprintf(w, `{"leader":"%s","term":%d,"nodes":%d}`,
			state.LeaderID, state.Term, len(state.Nodes))
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
			c.logger.Error().Err(err).Msg("HTTP server error")
		}
	}()
}
