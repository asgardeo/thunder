#!/usr/bin/env bash
set -euo pipefail

export NODE_ENV=production

# Copy your public/ folder into the standalone bundle
echo "Copying public folder to .next/standalone..."
cp -R public .next/standalone

# Ensure the on-disk _next directory exists under public
echo "Creating public/_next directory..."
mkdir -p .next/standalone/public/_next

# Copy Next’s static assets into public/_next/static
echo "Copying static assets to public/_next/static..."
cp -R .next/static .next/standalone/public/_next

# # Remove the standalone "next" package itself (not needed at runtime)
# echo "Removing standalone node_modules/next..."
# rm -rf .next/standalone/node_modules/next

# after copying public & static…
echo "Copying SSL certs…"
cp server.key server.cert .next/standalone/

# Copy the server.js file to the standalone directory
echo "Copying server.js to .next/standalone..."
cp server.js .next/standalone/

cd .next/standalone
# Install production dependencies
echo "Installing production dependencies..."
pnpm install --prod

echo "Post-build steps complete."
