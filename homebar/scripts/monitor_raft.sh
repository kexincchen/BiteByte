#!/bin/bash

# Simple script to monitor Raft cluster status
# Run with: ./monitor_raft.sh

while true; do
    echo "Checking Raft cluster status..."
    
    # Get cluster status
    status=$(curl -s http://localhost:8090/cluster/status)
    
    # Parse and display status (simplified for this example)
    echo "Cluster Status: $status"
    
    # Get node status
    nodes=$(curl -s http://localhost:8090/cluster/nodes)
    
    echo "Nodes: $nodes"
    echo ""
    
    # Wait before checking again
    sleep 5
done 