#!/bin/bash
# ----------------------------------------------------------------------------
# Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
#
# WSO2 LLC. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.
# ----------------------------------------------------------------------------

# Server port
SERVER_PORT=3000

set -e  # Exit immediately if a command exits with a non-zero status

# Kill known ports
function kill_port() {
    local port=$1
    lsof -ti tcp:$port | xargs kill -9 2>/dev/null || true
}

# Kill ports before binding
kill_port $SERVER_PORT

# Check if npx is available
if ! command -v npx &> /dev/null; then
    echo "❌ Error: npx is not installed. Please install Node.js and npm."
    exit 1
fi

# Check if certificates exist in dist folder
CERT_FILE="dist/server.cert"
KEY_FILE="dist/server.key"

echo "⚡ Starting React SDK Sample App Server on port $SERVER_PORT..."
echo ""
echo "📂 Serving static files from ./dist directory"
echo ""

# Run server using serve with HTTPS if certificates exist, otherwise HTTP
if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
    echo "🔐 Using HTTPS with SSL certificates"
    npx serve -s dist -l $SERVER_PORT --ssl-cert "$CERT_FILE" --ssl-key "$KEY_FILE" &
    SERVER_PID=$!
    PROTOCOL="https"
else
    echo "⚠️  SSL certificates not found. Running with HTTP"
    echo "    Run the build script to generate certificates"
    npx serve -s dist -l $SERVER_PORT &
    SERVER_PID=$!
    PROTOCOL="http"
fi

# Cleanup on Ctrl+C
trap 'echo -e "\n🛑 Stopping server..."; kill $SERVER_PID; exit' SIGINT

# Status
echo ""
echo "🚀 React SDK Sample App running at $PROTOCOL://localhost:$SERVER_PORT"
echo "Press Ctrl+C to stop the server."
echo ""

# Wait for background processes
wait $SERVER_PID
