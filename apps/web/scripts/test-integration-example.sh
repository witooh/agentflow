#!/bin/bash

# Example script to run integration tests with Supabase
# This assumes you have a local Supabase instance running

echo "🧪 Running Supabase Integration Tests"
echo "====================================="

# Check if Supabase is running locally
if ! curl -s http://127.0.0.1:54321/health > /dev/null; then
    echo "❌ Supabase local instance not running!"
    echo "Start it with: npx supabase start"
    exit 1
fi

echo "✅ Supabase instance detected"

# Set environment variables for integration tests
export RUN_INTEGRATION_TESTS=true
export NEXT_PUBLIC_SUPABASE_URL=http://127.0.0.1:54321
export NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0
export SUPABASE_URL=http://127.0.0.1:54321
export SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImV4cCI6MTk4MzgxMjk5Nn0.EGIM96RAZx35lJzdJsyH-qQwv8Hdp7fsn3W0YpN81IU

echo "🔧 Environment variables set"

# Run integration tests
echo "🚀 Running integration tests..."
npm run test:integration

echo "✅ Integration tests completed!"


