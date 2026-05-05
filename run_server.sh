#!/bin/bash

echo "=============================================="
echo " Markflow Studio - Server Restarter"
echo "=============================================="

echo "[1/2] Stopping existing server processes..."
# Kill go.exe and main.exe processes silently
taskkill //F //IM go.exe //IM main.exe 2>/dev/null
if [ $? -eq 0 ]; then
    echo "Successfully stopped old server."
else
    echo "No old server running."
fi

echo ""
echo "[2/2] Starting new server on port 8888..."
go run cmd/server/main.go
