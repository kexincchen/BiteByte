package raft

import (
	"time"
)

// NodeState represents the state of a Raft node
type NodeState string

const (
	// Follower state: the node receives log entries from a leader and votes
	Follower NodeState = "follower"
	// Candidate state: the node is campaigning to become a leader
	Candidate NodeState = "candidate"
	// Leader state: the node is the current leader, responsible for log replication
	Leader NodeState = "leader"
)

// Constants for Raft protocol
const (
	// Timeout ranges
	MinElectionTimeout  = 150 * time.Millisecond
	MaxElectionTimeout  = 300 * time.Millisecond
	HeartbeatInterval   = 50 * time.Millisecond
	RPCTimeout          = 100 * time.Millisecond
	MaxAppendEntries    = 100 // Maximum number of entries to send in a single AppendEntries RPC
	MaxLogEntriesBuffer = 1000
)

// LogEntry represents a single entry in the Raft log
type LogEntry struct {
	Index   uint64      // Position in the log
	Term    uint64      // Term when entry was received by leader
	Command interface{} // Command to be applied to the state machine
}

// OrderCommand represents a command to modify order and inventory state
type OrderCommand struct {
	Type           string                 // Type of command (e.g., "create_order", "update_inventory")
	OrderID        uint                   // ID of the order (if applicable)
	CustomerID     uint                   // Customer ID
	MerchantID     uint                   // Merchant ID
	OrderItems     []OrderItemCommand     // Items in the order
	AdditionalData map[string]interface{} // Additional command-specific data
}

// OrderItemCommand represents an item in an order command
type OrderItemCommand struct {
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// RequestVoteArgs represents the arguments for a RequestVote RPC
type RequestVoteArgs struct {
	Term         uint64 // Candidate's term
	CandidateID  string // Candidate requesting vote
	LastLogIndex uint64 // Index of candidate's last log entry
	LastLogTerm  uint64 // Term of candidate's last log entry
}

// RequestVoteReply represents the result of a RequestVote RPC
type RequestVoteReply struct {
	Term        uint64 // Current term, for candidate to update itself
	VoteGranted bool   // True if candidate received vote
}

// AppendEntriesArgs represents the arguments for an AppendEntries RPC
type AppendEntriesArgs struct {
	Term         uint64     // Leader's term
	LeaderID     string     // So follower can redirect clients
	PrevLogIndex uint64     // Index of log entry immediately preceding new ones
	PrevLogTerm  uint64     // Term of prevLogIndex entry
	Entries      []LogEntry // Log entries to store (empty for heartbeat)
	LeaderCommit uint64     // Leader's commitIndex
}

// AppendEntriesReply represents the result of an AppendEntries RPC
type AppendEntriesReply struct {
	Term          uint64 // Current term, for leader to update itself
	Success       bool   // True if follower contained entry matching prevLogIndex and prevLogTerm
	ConflictTerm  uint64 // Term of the conflicting entry
	ConflictIndex uint64 // First index of the conflicting term
}
