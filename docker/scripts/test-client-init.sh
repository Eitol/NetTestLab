#!/bin/bash

# Test Client Initialization Script
# Creates a tcpdump wrapper for testing purposes

set -e

echo "Initializing test client with tcpdump support..."

# Create a simple tcpdump wrapper for the test client environment
cat > /usr/local/bin/tcpdump << 'EOF'
#!/bin/bash

# Simple tcpdump wrapper for test client
# This is just for testing that tcpdump command is available

echo "tcpdump wrapper for test environment - version 4.99.1"

# Handle basic options
case "$1" in
    --version|-V)
        echo "tcpdump version 4.99.1"
        echo "libpcap version 1.10.1"
        exit 0
        ;;
    --help|-h)
        echo "Usage: tcpdump [options] [expression]"
        echo "This is a test wrapper - real tcpdump functionality handled by NetTestLab"
        exit 0
        ;;
    *)
        echo "tcpdump test wrapper: would capture with args: $*"
        exit 0
        ;;
esac
EOF

chmod +x /usr/local/bin/tcpdump

echo "✓ tcpdump wrapper created for test client"

# Test the wrapper
if command -v tcpdump >/dev/null 2>&1; then
    echo "✓ tcpdump is available in test client"
    tcpdump --version 2>&1 | head -1
else
    echo "✗ tcpdump wrapper failed"
    exit 1
fi

echo "Test client initialization complete"