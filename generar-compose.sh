#!/bin/bash

if [ -z "$1" ]; then
  echo "Error: file path not specified"
  echo "Usage: $0 <output-file> <n-clients>"
  exit 1
fi

# 1° param
OUTPUT_FILE="$1"

# 2° param, if not set, uses 1
NUM_CLIENTS="${2:-1}"

if ! [[ "$NUM_CLIENTS" =~ ^[0-9]+$ ]]; then
  echo "Error: N° of clients must be a positive integer"
  echo "Usage: $0 <output-file> <n-clients>"
  exit 1
fi

echo "Generating docker compose with $NUM_CLIENTS clients: $OUTPUT_FILE"

python3 generar-compose.py $OUTPUT_FILE $NUM_CLIENTS