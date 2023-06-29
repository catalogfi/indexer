#!/bin/bash
count=1
GREEN='\033[0;32m'
RED='\033[0;31m'
while true; do
  echo -e "${GREEN}[+]Running: ${count}"
  go run ./cmd/peer/peer.go &
  pid=$!
  sleep 10
  echo -e "${RED}[-]Killing: ${count}"
  kill $pid
  ((count++))
done
