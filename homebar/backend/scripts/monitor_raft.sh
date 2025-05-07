#!/bin/bash

# Script to monitor the entire Raft cluster
# Run with: ./monitor_raft.sh

# Define port range for the coordinator nodes
START_PORT=8091
END_PORT=8093  # Adjust based on your cluster size

# ANSI color codes for better visualization
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Clear terminal
clear

while true; do
    clear
    echo -e "${BOLD}${BLUE}===== Raft Cluster Status =====${NC}"
    echo -e "${CYAN}Time: $(date)${NC}"
    echo ""
    
    # Track active nodes and leader information
    active_nodes=0
    leader_id=""
    leader_port=""
    
    # First, get the global cluster state from any responding node
    for port in $(seq $START_PORT $END_PORT); do
        status=$(curl -s --connect-timeout 1 http://localhost:$port/cluster/status)
        if [ $? -eq 0 ] && [ ! -z "$status" ]; then
            # Extract leader ID directly from the cluster status
            leader_id=$(echo $status | grep -o '"leader":"[^"]*"' | cut -d'"' -f4)
            if [ ! -z "$leader_id" ]; then
                # Calculate leader port based on ID
                leader_port=$((START_PORT + $(echo $leader_id | sed 's/[^0-9]*//g') - 1))
                break
            fi
        fi
    done
    
    # Check each coordinator node
    for port in $(seq $START_PORT $END_PORT); do
        node_id=$((port - START_PORT + 1))
        echo -e "${BOLD}Node $node_id${NC} (port $port):"
        
        # Get cluster status from this node
        status=$(curl -s --connect-timeout 1 http://localhost:$port/cluster/status)
        
        if [ $? -eq 0 ] && [ ! -z "$status" ]; then
            active_nodes=$((active_nodes + 1))
            
            # Extract term and node count
            current_term=$(echo $status | grep -o '"term":[0-9]*' | cut -d':' -f2)
            node_count=$(echo $status | grep -o '"nodes":[0-9]*' | cut -d':' -f2)
            
            # Determine if this node is the leader
            if [[ "$node_id" == "$leader_id" ]]; then
                echo -e "  State: ${GREEN}${BOLD}LEADER${NC} (Term: $current_term)"
            else
                echo -e "  State: ${BLUE}Follower${NC} (Term: $current_term)"
            fi
            
            echo -e "  Sees $node_count active node(s)"
            
            # Get node details - try to format the output
            nodes=$(curl -s --connect-timeout 1 http://localhost:$port/cluster/nodes)
            if [ ! -z "$nodes" ]; then
                echo -e "  Peers: $nodes"
            fi
            
        else
            echo -e "  ${RED}⚠️ Node not responding${NC}"
        fi
        
        echo ""
    done
    
    echo -e "${BOLD}Cluster Summary:${NC}"
    echo -e "  Active Nodes: ${active_nodes}/$((END_PORT - START_PORT + 1))"
    
    if [ ! -z "$leader_id" ]; then
        echo -e "  Current Leader: ${GREEN}Node $leader_id${NC} (port $leader_port)"
    else
        echo -e "  Current Leader: ${RED}No leader detected${NC}"
    fi
    
    echo -e "${BLUE}================================${NC}"
    
    # Display help information
    echo -e "${YELLOW}Press Ctrl+C to exit the monitor${NC}"
    
    # Wait before checking again
    sleep 5
done