package raft

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// RaftNode represents a node in the Raft cluster
type RaftNode struct {
	// Node state
	state       NodeState
	id          string
	peers       map[string]*RaftPeer
	peerAddrs   map[string]string
	currentTerm uint64
	votedFor    string
	log         []LogEntry
	commitIndex uint64
	lastApplied uint64

	// Leader state
	nextIndex  map[string]uint64
	matchIndex map[string]uint64

	// Communication and synchronization
	applyCh chan LogEntry // Channel to send committed log entries to state machine
	mu      sync.Mutex    // Protects concurrent access to node state

	// Timers
	electionTimer  *time.Timer
	heartbeatTimer *time.Timer

	// Configuration
	heartbeatInterval time.Duration

	// State machine application function (executes committed commands)
	applyCommand func(cmd interface{}) error

	// For logging and debugging
	logger   *log.Logger
	leaderID string

	// Add storage field
	storage Storage
}

// NewRaftNode creates a new Raft node with the given configuration
func NewRaftNode(id string, peers []string, peerAddrs map[string]string, applyCh chan LogEntry, applyCommand func(cmd interface{}) error, logger *log.Logger) *RaftNode {
	node := &RaftNode{
		id:                id,
		peers:             make(map[string]*RaftPeer),
		peerAddrs:         peerAddrs,
		log:               []LogEntry{{Term: 0, Index: 0}}, // Start with a dummy entry
		applyCh:           applyCh,
		currentTerm:       0,
		votedFor:          "",
		commitIndex:       0,
		lastApplied:       0,
		nextIndex:         make(map[string]uint64),
		matchIndex:        make(map[string]uint64),
		applyCommand:      applyCommand,
		heartbeatInterval: HeartbeatInterval,
		logger:            logger,
	}

	// Initialize storage
	storageDir := os.Getenv("RAFT_STORAGE_DIR")
	storage, err := NewFileStorage(id, storageDir)
	if err != nil {
		logger.Printf("Failed to initialize storage: %v", err)
		// Continue with in-memory only as fallback
	} else {
		node.storage = storage

		// Load persistent state
		term, votedFor, lastApplied, err := storage.LoadState()
		if err != nil {
			logger.Printf("Failed to load state: %v", err)
		} else {
			node.currentTerm = term
			node.votedFor = votedFor
			node.lastApplied = lastApplied
			node.commitIndex = max(node.commitIndex, lastApplied)
		}

		// Load log
		log, err := storage.LoadLog()
		if err != nil {
			logger.Printf("Failed to load log: %v", err)
		} else if len(log) > 0 {
			node.log = log
		}
	}

	// Initialize peer connections
	for _, peerID := range peers {
		if peerID != id { // Don't add self as peer
			node.peers[peerID] = &RaftPeer{
				id:     peerID,
				client: nil, // Will be initialized when starting the node
			}
		}
	}

	return node
}

// Start initializes the Raft node and begins operation
func (n *RaftNode) Start(ctx context.Context) error {
	n.logger.Printf("Starting Raft node %s", n.id)

	// Initialize peer connections
	for id, peer := range n.peers {
		addr := n.peerAddrs[id]
		if addr == "" {
			addr = fmt.Sprintf("http://localhost:808%s/raft", id)
		}
		client, err := NewRaftClient(id, addr)
		if err != nil {
			return fmt.Errorf("failed to connect to peer %s: %w", id, err)
		}
		peer.client = client
	}

	// Become follower at the term we have just loaded
	n.becomeFollower(n.currentTerm)

	// Start the main loop
	go n.run(ctx)

	return nil
}

// run is the main loop of the Raft node
func (n *RaftNode) run(ctx context.Context) {
	// Initialize timers if they haven't been initialized yet
	if n.electionTimer == nil {
		timeout := MinElectionTimeout + time.Duration(rand.Int63n(int64(MaxElectionTimeout-MinElectionTimeout)))
		n.electionTimer = time.NewTimer(timeout)
	}

	if n.heartbeatTimer == nil {
		n.heartbeatTimer = time.NewTimer(n.heartbeatInterval)
	}

	for {
		select {
		case <-ctx.Done():
			n.logger.Printf("Shutting down Raft node %s", n.id)
			return

		case <-n.electionTimer.C:
			n.mu.Lock()
			if n.state != Leader {
				n.startElection()
			}
			n.mu.Unlock()

		case <-n.heartbeatTimer.C:
			n.mu.Lock()
			if n.state == Leader {
				n.sendHeartbeats()
				n.heartbeatTimer.Reset(n.heartbeatInterval)
			}
			n.mu.Unlock()
		}

		// Apply committed entries to state machine
		n.applyCommittedEntries()
	}
}

// becomeFollower transitions this node to follower state
func (n *RaftNode) becomeFollower(term uint64) {
	oldTerm := n.currentTerm
	if term > n.currentTerm {
		n.currentTerm = term
	}

	n.state = Follower

	if term > oldTerm {
		n.votedFor = ""
	}
	// Persist state to storage
	n.persistState()

	// Reset election timer with random timeout
	timeout := MinElectionTimeout + time.Duration(rand.Int63n(int64(MaxElectionTimeout-MinElectionTimeout)))
	if n.electionTimer == nil {
		n.electionTimer = time.NewTimer(timeout)
	} else {
		n.electionTimer.Reset(timeout)
	}
}

// becomeCandidate transitions this node to candidate state
func (n *RaftNode) becomeCandidate() {
	n.state = Candidate
	n.currentTerm++
	n.votedFor = n.id

	// Persist state to storage
	n.persistState()

	// Reset election timer
	timeout := MinElectionTimeout + time.Duration(rand.Int63n(int64(MaxElectionTimeout-MinElectionTimeout)))
	n.electionTimer.Reset(timeout)
}

// becomeLeader transitions this node to leader state
func (n *RaftNode) becomeLeader() {
	n.state = Leader
	n.leaderID = n.id

	n.logger.Printf("👑 Node %s becomes LEADER for term %d", n.id, n.currentTerm)

	// Initialize nextIndex and matchIndex
	lastLogIndex := uint64(len(n.log) - 1)
	for peerID := range n.peers {
		n.nextIndex[peerID] = lastLogIndex + 1
		n.matchIndex[peerID] = 0
	}

	// Start sending heartbeats
	if n.heartbeatTimer == nil {
		n.heartbeatTimer = time.NewTimer(n.heartbeatInterval)
	} else {
		n.heartbeatTimer.Reset(n.heartbeatInterval)
	}
}

// startElection initiates a new election
func (n *RaftNode) startElection() {
	n.becomeCandidate()

	// Vote for self
	votesReceived := 1

	participants := 1
	n.logger.Printf("⏳ Node %s starts election for term %d", n.id, n.currentTerm)

	// Prepare RequestVote arguments
	lastLogIndex := uint64(len(n.log) - 1)
	lastLogTerm := n.log[lastLogIndex].Term

	args := RequestVoteArgs{
		Term:         n.currentTerm,
		CandidateID:  n.id,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}

	// Ask for votes from all peers
	var votesMu sync.Mutex
	var wg sync.WaitGroup

	for _, peer := range n.peers {
		wg.Add(1)
		go func(p *RaftPeer) {
			defer wg.Done()

			var reply RequestVoteReply
			if err := p.client.RequestVote(args, &reply); err != nil {
				n.logger.Printf("Error requesting vote from %s: %v", p.id, err)
				n.logger.Printf("⚠️  Node %s vote request to %s failed: %v", n.id, p.id, err)
				return
			}

			n.mu.Lock()
			defer n.mu.Unlock()

			// If we've already moved on to a new term, ignore this response
			if n.state != Candidate || n.currentTerm != args.Term {
				return
			}

			// If the peer has a higher term, become follower
			if reply.Term > n.currentTerm {
				n.becomeFollower(reply.Term)
				return
			}

			// Count vote if granted
			if reply.VoteGranted {
				votesMu.Lock()
				votesReceived++
				participants++
				votesMu.Unlock()

				// Check if we have majority
				if votesReceived > (len(n.peers)+1)/2 {
					n.becomeLeader()
				}

				n.logger.Printf("✅  Node %s got vote from %s (%d/%d)",
					n.id, p.id, votesReceived, participants)

			}
		}(peer)
	}
}

// sendHeartbeats sends heartbeats to all peers
func (n *RaftNode) sendHeartbeats() {
	for _, peer := range n.peers {
		go n.sendAppendEntries(peer)
	}
}

// sendAppendEntries sends an AppendEntries RPC to a peer
func (n *RaftNode) sendAppendEntries(peer *RaftPeer) {
	nextIdx := n.nextIndex[peer.id]
	prevLogIndex := nextIdx - 1
	prevLogTerm := uint64(0)

	if prevLogIndex > 0 && prevLogIndex < uint64(len(n.log)) {
		prevLogTerm = n.log[prevLogIndex].Term
	}

	// Get entries to send
	var entries []LogEntry
	if nextIdx < uint64(len(n.log)) {
		entries = n.log[nextIdx:]
		if len(entries) > MaxAppendEntries {
			entries = entries[:MaxAppendEntries]
		}
	}

	args := AppendEntriesArgs{
		Term:         n.currentTerm,
		LeaderID:     n.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: n.commitIndex,
	}

	var reply AppendEntriesReply
	if err := peer.client.AppendEntries(args, &reply); err != nil {
		fmt.Printf("Error sending AppendEntries to %s: %v\n", peer.id, err)
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// If we're no longer the leader or term has changed, ignore response
	if n.state != Leader || n.currentTerm != args.Term {
		return
	}

	// If peer has higher term, become follower
	if reply.Term > n.currentTerm {
		n.becomeFollower(reply.Term)
		return
	}

	if reply.Success {
		// fmt.Printf("AppendEntries to %s successful\n", peer.id)
		// Update nextIndex and matchIndex for successful append
		n.matchIndex[peer.id] = prevLogIndex + uint64(len(entries))
		n.nextIndex[peer.id] = n.matchIndex[peer.id] + 1

		// Check if we can commit more entries
		n.updateCommitIndex()
	} else {
		fmt.Printf("AppendEntries to %s failed\n", peer.id)
		// If append failed, decrement nextIndex and retry
		if reply.ConflictTerm > 0 {
			// Fast backtracking using conflict information
			conflictTermStartIndex := uint64(0)
			// Find the first index of conflicting term in our log
			for i := prevLogIndex; i > 0; i-- {
				if i < uint64(len(n.log)) && n.log[i].Term == reply.ConflictTerm {
					conflictTermStartIndex = i
					break
				}
			}

			if conflictTermStartIndex > 0 {
				// We found an entry with the conflict term, try the next index
				n.nextIndex[peer.id] = conflictTermStartIndex + 1
			} else {
				// We don't have the conflict term, go to the first index of that term
				n.nextIndex[peer.id] = reply.ConflictIndex
			}
		} else {
			// Simpler backtracking if conflict info not provided
			n.nextIndex[peer.id] = max(1, n.nextIndex[peer.id]-1)
		}
	}
}

// updateCommitIndex updates the commit index based on matchIndex values
func (n *RaftNode) updateCommitIndex() {
	// Find the highest index that is replicated to a majority of nodes
	for i := n.commitIndex + 1; i < uint64(len(n.log)); i++ {
		// Only consider entries from current term
		if n.log[i].Term != n.currentTerm {
			continue
		}

		count := 1 // Count self
		for _, matchIdx := range n.matchIndex {
			if matchIdx >= i {
				count++
			}
		}

		// Check if we have a majority
		if count > (len(n.peers)+1)/2 {
			n.commitIndex = i
		} else {
			break
		}
	}
}

// applyCommittedEntries applies any newly committed entries to the state machine
func (n *RaftNode) applyCommittedEntries() {
	n.mu.Lock()
	defer n.mu.Unlock()

	for n.lastApplied < n.commitIndex {
		// Record previous value to check for changes
		prevLastApplied := n.lastApplied

		for n.lastApplied < n.commitIndex {
			n.lastApplied++
			entry := n.log[n.lastApplied]
			n.applyCh <- entry
		}

		// If lastApplied changed, persist state
		if prevLastApplied != n.lastApplied {
			n.persistState()
		}

		// n.lastApplied++

		// // Apply the command to the state machine
		// entry := n.log[n.lastApplied]
		// n.applyCh <- entry

		// if n.applyCommand != nil {
		// 	if err := n.applyCommand(entry.Command); err != nil {
		// 		n.logger.Printf("Error applying command: %v", err)
		// 	}
		// }
	}
}

// Submit adds a new command to the log (called by clients)
func (n *RaftNode) Submit(command interface{}) (uint64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// If not the leader, reject the command
	if n.state != Leader {
		return 0, fmt.Errorf("not the leader")
	}

	// Append to log
	index := uint64(len(n.log))
	entry := LogEntry{
		Index:   index,
		Term:    n.currentTerm,
		Command: command,
	}

	n.log = append(n.log, entry)
	n.logger.Printf("🔄 Node %s submitted command at index %d", n.id, index)

	// Persist log entry
	n.persistLog(index)

	// Send the new entry to all peers immediately
	go n.sendHeartbeats()

	return index, nil
}

// RequestVote handles a RequestVote RPC from another node
func (n *RaftNode) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Update term if necessary
	termChanged := false
	if args.Term > n.currentTerm {
		n.becomeFollower(args.Term)
		termChanged = true
	}

	reply.Term = n.currentTerm
	reply.VoteGranted = false

	// Check if we can vote for this candidate
	if args.Term < n.currentTerm {
		// Reject vote if candidate's term is smaller
		return nil
	}

	// If we haven't voted yet or already voted for this candidate
	if n.votedFor == "" || n.votedFor == args.CandidateID {
		// Check if candidate's log is at least as up-to-date as ours
		lastLogIndex := uint64(len(n.log) - 1)
		lastLogTerm := n.log[lastLogIndex].Term

		if args.LastLogTerm > lastLogTerm ||
			(args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex) {
			// Grant vote
			prevVotedFor := n.votedFor
			n.votedFor = args.CandidateID
			reply.VoteGranted = true

			// Persist state if changed
			if prevVotedFor != n.votedFor || termChanged {
				n.persistState()
			}

			// Reset election timer since we voted
			timeout := MinElectionTimeout + time.Duration(rand.Int63n(int64(MaxElectionTimeout-MinElectionTimeout)))
			n.electionTimer.Reset(timeout)
		}
	}

	return nil
}

// AppendEntries handles an AppendEntries RPC from the leader
func (n *RaftNode) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	reply.Success = false
	reply.Term = n.currentTerm

	// If term is smaller than current term, reject
	if args.Term < n.currentTerm {
		return nil
	}

	// If we get a heartbeat from a leader with equal or higher term
	if args.Term >= n.currentTerm {
		prevTerm := n.currentTerm
		n.becomeFollower(args.Term)
		n.leaderID = args.LeaderID

		// If term changed, persist state
		if prevTerm != args.Term {
			n.persistState()
		}

		// Reset election timer on valid heartbeat
		timeout := MinElectionTimeout + time.Duration(rand.Int63n(int64(MaxElectionTimeout-MinElectionTimeout)))
		n.electionTimer.Reset(timeout)
	}

	// Check if we have the previous log entry
	if args.PrevLogIndex > 0 {
		if args.PrevLogIndex >= uint64(len(n.log)) {
			// We don't have this entry, provide hint
			reply.ConflictIndex = uint64(len(n.log))
			reply.ConflictTerm = 0
			return nil
		}

		if n.log[args.PrevLogIndex].Term != args.PrevLogTerm {
			// We have a conflicting entry, provide term info
			reply.ConflictTerm = n.log[args.PrevLogIndex].Term

			// Find first index with conflicting term
			for i := uint64(1); i < uint64(len(n.log)); i++ {
				if n.log[i].Term == reply.ConflictTerm {
					reply.ConflictIndex = i
					break
				}
			}
			return nil
		}
	}

	// Success case: we have the matching previous entry
	reply.Success = true

	// Append any new entries
	if len(args.Entries) > 0 {
		nextIdx := args.PrevLogIndex + 1
		logChanged := false
		persistIndex := uint64(len(n.log))

		// Handle new entries
		for i, entry := range args.Entries {
			if nextIdx+uint64(i) < uint64(len(n.log)) {
				// Entry exists, check if terms match
				if n.log[nextIdx+uint64(i)].Term != entry.Term {
					// Terms don't match, truncate log and append new entries
					n.log = n.log[:nextIdx+uint64(i)]
					n.log = append(n.log, args.Entries[i:]...)
					persistIndex = nextIdx + uint64(i)
					logChanged = true
					break
				}
			} else {
				// Reached end of existing log, append remaining entries
				n.log = append(n.log, args.Entries[i:]...)
				persistIndex = nextIdx + uint64(i)
				logChanged = true
				break
			}
		}

		// Persist log if changed
		if logChanged && n.storage != nil {
			n.persistLog(persistIndex)
		}
	}

	// Update commit index if needed
	if args.LeaderCommit > n.commitIndex {
		n.commitIndex = min(args.LeaderCommit, uint64(len(n.log)-1))
	}

	return nil
}

// IsLeader reports whether this node is the current leader.
func (n *RaftNode) IsLeader() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.state == Leader
}

// LeaderID returns current known leader id (may be empty).
func (n *RaftNode) LeaderID() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.leaderID
}

func (n *RaftNode) ID() string {
	return n.id
}

// Helper functions for min/max that work with uint64
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// Add these methods to access unexported fields
func (n *RaftNode) GetMutex() *sync.Mutex {
	return &n.mu
}

func (n *RaftNode) GetLastApplied() uint64 {
	return n.lastApplied
}

func (n *RaftNode) SetLastApplied(value uint64) {
	n.lastApplied = value
}

// Add this method to update lastApplied
func (n *RaftNode) UpdateLastApplied(index uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.lastApplied = index
}

// Add method to persist Raft state
func (n *RaftNode) persistState() {
	if n.storage != nil {
		if err := n.storage.SaveState(n.currentTerm, n.votedFor, n.lastApplied); err != nil {
			n.logger.Printf("Failed to persist state: %v", err)
		}
	}
}

// Add method to persist log entries
func (n *RaftNode) persistLog(startIndex uint64) {
	if n.storage == nil || startIndex >= uint64(len(n.log)) {
		return
	}

	entries := n.log[startIndex:]
	if err := n.storage.AppendLog(entries); err != nil {
		n.logger.Printf("Failed to persist log entries: %v", err)
	}
}
