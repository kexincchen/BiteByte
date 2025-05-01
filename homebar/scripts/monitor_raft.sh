#!/bin/bash

# Script to monitor the entire Raft cluster
# Run with: ./monitor_raft.sh

# Define port range for the coordinator nodes
START_PORT=8091
END_PORT=8093  # Adjust based on your cluster size

while true; do
    echo "===== Raft Cluster Status ====="
    echo "Time: $(date)"
    
    # Track active nodes
    active_nodes=0
    
    # Check each coordinator node
    for port in $(seq $START_PORT $END_PORT); do
        echo "Checking node on port $port..."
        
        # Get cluster status from this node
        status=$(curl -s --connect-timeout 1 http://localhost:$port/cluster/status)
        
        if [ $? -eq 0 ] && [ ! -z "$status" ]; then
            echo "  Status: $status"
            active_nodes=$((active_nodes + 1))
            
            # Get node details
            nodes=$(curl -s --connect-timeout 1 http://localhost:$port/cluster/nodes)
            echo "  Nodes: $nodes"
        else
            echo "  ⚠️ Node not responding"
        fi
        echo ""
    done
    
    echo "Active Nodes: $active_nodes/$((END_PORT - START_PORT + 1))"
    echo "================================"
    echo ""
    
    # Wait before checking again
    sleep 5
done 