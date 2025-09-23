#!/usr/bin/env bash
set -euo pipefail

CLIENT_NAME="${1:-fedup}"
SERVER_NAME="${2:-fedupd}"

echo "Client binary name: $CLIENT_NAME"
echo "Server binary name: $SERVER_NAME"

echo "==> Building client"
go build -v -o "$CLIENT_NAME" "./client"
echo "==> Building server"
go build -v -o "$SERVER_NAME" "./server"
