#!/bin/bash

echo "==================================================================="
echo "  CohortBridge Two-Party Network Demo"
echo "==================================================================="
echo ""
echo "This demo shows how to run CohortBridge agent in true two-party mode"
echo "where the sender and receiver communicate over the network."
echo ""
echo "Setup Instructions:"
echo "-------------------"
echo ""
echo "DEVICE 1 (Receiver):"
echo "  1. Ensure data/patients_a_small2.csv exists"
echo "  2. Run: ./agent.exe -mode=receiver -config=config_receiver.yaml"
echo "     This will:"
echo "     - Create temp-receiver/ directory"
echo "     - Tokenize local data"
echo "     - Listen on port 8080 for sender"
echo "     - Exchange tokenized data"
echo "     - Compute intersection"
echo "     - Receive final results"
echo ""
echo "DEVICE 2 (Sender):"
echo "  1. Ensure data/patients_b_small2.csv exists"
echo "  2. Wait for receiver to start listening"
echo "  3. Run: ./agent.exe -mode=sender -config=config_sender.yaml"
echo "     This will:"
echo "     - Create temp-sender/ directory"
echo "     - Tokenize local data"
echo "     - Connect to receiver at 127.0.0.1:8080"
echo "     - Exchange tokenized data"
echo "     - Compute intersection"
echo "     - Send final results to receiver"
echo ""
echo "Expected Flow:"
echo "1. Receiver starts and listens on port 8080"
echo "2. Sender connects and sends its tokenized data"
echo "3. Receiver sends its tokenized data back"
echo "4. Both compute intersection locally"
echo "5. Sender sends intersection results to receiver"
echo ""
echo "Results:"
echo "- Receiver: temp-receiver/intersection.csv + temp-receiver/sender_intersection.csv"
echo "- Sender: temp-sender/intersection.csv"
echo ""
echo "For testing on same machine:"
echo "----------------------------"
echo ""

# Function to run receiver in background
run_receiver() {
    echo "🔵 Starting receiver (background)..."
    ./agent.exe -mode=receiver -config=config_receiver.yaml &
    RECEIVER_PID=$!
    echo "   Receiver PID: $RECEIVER_PID"
    sleep 3  # Give receiver time to start
}

# Function to run sender
run_sender() {
    echo "🟡 Starting sender..."
    ./agent.exe -mode=sender -config=config_sender.yaml
    SENDER_EXIT=$?
    return $SENDER_EXIT
}

# Function to cleanup
cleanup() {
    if [ ! -z "$RECEIVER_PID" ]; then
        echo "🛑 Stopping receiver (PID: $RECEIVER_PID)..."
        kill $RECEIVER_PID 2>/dev/null || true
        wait $RECEIVER_PID 2>/dev/null || true
    fi
}

# Check if we should run the demo
if [ "$1" = "run" ]; then
    echo "🚀 Running two-party demo on same machine..."
    echo ""
    
    # Clean up any existing temp directories
    rm -rf temp-receiver temp-sender
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Start receiver
    run_receiver
    
    # Start sender
    run_sender
    RESULT=$?
    
    # Show results
    echo ""
    echo "==================================================================="
    echo "  Demo Results"
    echo "==================================================================="
    
    if [ $RESULT -eq 0 ]; then
        echo "✅ Demo completed successfully!"
        echo ""
        echo "Receiver files:"
        ls -la temp-receiver/ 2>/dev/null || echo "  (no temp-receiver directory)"
        echo ""
        echo "Sender files:"
        ls -la temp-sender/ 2>/dev/null || echo "  (no temp-sender directory)"
        echo ""
        
        if [ -f "temp-receiver/intersection.csv" ]; then
            echo "Receiver intersection results:"
            cat temp-receiver/intersection.csv
        fi
        
        if [ -f "temp-sender/intersection.csv" ]; then
            echo "Sender intersection results:"
            cat temp-sender/intersection.csv
        fi
    else
        echo "❌ Demo failed with exit code: $RESULT"
    fi
    
    cleanup
else
    echo "To run the demo on the same machine: $0 run"
    echo ""
    echo "To run on separate devices:"
    echo "  Device 1: ./agent.exe -mode=receiver -config=config_receiver.yaml"
    echo "  Device 2: ./agent.exe -mode=sender -config=config_sender.yaml"
fi 