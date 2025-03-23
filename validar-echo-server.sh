#!/bin/bash

SERVER_NAME="server"
SERVER_PORT=12345
NETWORK_NAME="testing_net"
TEST_MESSAGE="Is this my echo?"

# Reuse busybox image (used in client Dockerfile)
IMAGE="busybox"

RESPONSE=$(docker run --rm --network $NETWORK_NAME $IMAGE sh -c "echo '$TEST_MESSAGE' | nc $SERVER_NAME $SERVER_PORT")

if [ "$RESPONSE" == "$TEST_MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
