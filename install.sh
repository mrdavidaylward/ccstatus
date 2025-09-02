#!/bin/bash

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root for system-wide install
check_permissions() {
    if [[ $EUID -eq 0 ]]; then
        log_error "Don't run this script as root. It will use sudo when needed."
        exit 1
    fi
}

# Install Go if not present
install_go() {
    if ! command -v go &> /dev/null; then
        log_warn "Go is not installed. Installing Go..."
        
        # Detect architecture
        ARCH=$(uname -m)
        case $ARCH in
            x86_64) GOARCH="amd64" ;;
            aarch64|arm64) GOARCH="arm64" ;;
            armv6l) GOARCH="armv6l" ;;
            armv7l) GOARCH="armv7l" ;;
            *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        
        # Download and install Go
        GO_VERSION="1.21.5"
        GO_TAR="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
        
        log_info "Downloading Go ${GO_VERSION} for ${GOARCH}..."
        cd /tmp
        curl -LO "https://golang.org/dl/${GO_TAR}"
        
        log_info "Installing Go to /usr/local/go..."
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "${GO_TAR}"
        
        # Add Go to PATH
        if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        fi
        
        export PATH=$PATH:/usr/local/go/bin
        log_success "Go installed successfully"
    else
        log_success "Go is already installed: $(go version)"
    fi
}

# Install Node.js and npm if not present (needed for ccusage)
install_nodejs() {
    if ! command -v npm &> /dev/null; then
        log_warn "npm is not installed. Installing Node.js and npm..."
        
        # Use NodeSource repository for latest Node.js
        curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
        sudo apt-get install -y nodejs
        
        log_success "Node.js and npm installed successfully"
    else
        log_success "npm is already installed: $(npm --version)"
    fi
}

# Install ccusage CLI tool
install_ccusage() {
    if ! command -v ccusage &> /dev/null; then
        log_info "Installing ccusage CLI tool..."
        sudo npm install -g ccusage
        log_success "ccusage installed successfully"
    else
        log_info "ccusage is already installed: $(ccusage --version 2>/dev/null || echo 'version unknown')"
        log_info "Updating ccusage to latest version..."
        sudo npm update -g ccusage
        log_success "ccusage updated successfully"
    fi
}

# Build ccstatus binary
build_ccstatus() {
    log_info "Building ccstatus binary..."
    
    if [[ ! -f "main.go" ]]; then
        log_error "main.go not found. Please run this script from the ccstatus directory."
        exit 1
    fi
    
    # Initialize Go module if not exists
    if [[ ! -f "go.mod" ]]; then
        log_info "Initializing Go module..."
        go mod init ccstatus
    fi
    
    # Build the binary
    go build -ldflags "-s -w" -o ccstatus main.go
    
    if [[ ! -f "ccstatus" ]]; then
        log_error "Failed to build ccstatus binary"
        exit 1
    fi
    
    log_success "ccstatus binary built successfully"
}

# Install ccstatus to system PATH
install_binary() {
    log_info "Installing ccstatus to /usr/local/bin..."
    
    # Make binary executable
    chmod +x ccstatus
    
    # Copy to system location
    sudo cp ccstatus /usr/local/bin/ccstatus
    
    # Verify installation
    if command -v ccstatus &> /dev/null; then
        log_success "ccstatus installed to /usr/local/bin/ccstatus"
    else
        log_error "Failed to install ccstatus binary"
        exit 1
    fi
}

# Create Claude config directory
setup_claude_config() {
    CLAUDE_DIR="$HOME/.claude"
    
    log_info "Setting up Claude configuration directory..."
    
    # Create directory if it doesn't exist
    mkdir -p "$CLAUDE_DIR"
    
    # Create settings.json if it doesn't exist
    SETTINGS_FILE="$CLAUDE_DIR/settings.json"
    if [[ ! -f "$SETTINGS_FILE" ]]; then
        log_info "Creating initial Claude settings.json..."
        cat > "$SETTINGS_FILE" << 'EOF'
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "model": "sonnet",
  "statusLine": {
    "type": "command",
    "command": "/usr/local/bin/ccstatus"
  }
}
EOF
        log_success "Created $SETTINGS_FILE"
    else
        log_info "Claude settings.json already exists"
    fi
}

# Update Claude status line configuration
update_statusline_config() {
    SETTINGS_FILE="$HOME/.claude/settings.json"
    
    log_info "Updating Claude status line configuration..."
    
    # Backup original settings
    cp "$SETTINGS_FILE" "$SETTINGS_FILE.backup.$(date +%Y%m%d_%H%M%S)"
    log_info "Backed up original settings to ${SETTINGS_FILE}.backup.*"
    
    # Update status line configuration using jq if available, otherwise manual
    if command -v jq &> /dev/null; then
        # Use jq for safe JSON manipulation
        jq '.statusLine.type = "command" | .statusLine.command = "/usr/local/bin/ccstatus"' "$SETTINGS_FILE" > "$SETTINGS_FILE.tmp"
        mv "$SETTINGS_FILE.tmp" "$SETTINGS_FILE"
        log_success "Updated status line configuration using jq"
    else
        # Manual JSON update (less robust but works)
        log_warn "jq not found, using manual JSON update"
        
        # Read current settings
        if grep -q '"statusLine"' "$SETTINGS_FILE"; then
            # Replace existing statusLine
            sed -i 's|"statusLine".*{[^}]*}|"statusLine": {\n    "type": "command",\n    "command": "/usr/local/bin/ccstatus"\n  }|' "$SETTINGS_FILE"
        else
            # Add statusLine before closing brace
            sed -i 's|^}$|,\n  "statusLine": {\n    "type": "command",\n    "command": "/usr/local/bin/ccstatus"\n  }\n}|' "$SETTINGS_FILE"
        fi
        log_success "Updated status line configuration manually"
    fi
}

# Setup initial tracking files
setup_tracking_files() {
    CLAUDE_DIR="$HOME/.claude"
    
    log_info "Setting up tracking files..."
    
    # Create latency tracking file with sample data
    LATENCY_FILE="$CLAUDE_DIR/latency.txt"
    if [[ ! -f "$LATENCY_FILE" ]]; then
        cat > "$LATENCY_FILE" << 'EOF'
1250.5
980.2
15
EOF
        log_success "Created latency tracking file: $LATENCY_FILE"
    fi
    
    # Create session start time file
    SESSION_START_FILE="$CLAUDE_DIR/session_start"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$SESSION_START_FILE"
    log_success "Created session start file: $SESSION_START_FILE"
    
    # Create current session ID file
    SESSION_FILE="$CLAUDE_DIR/current_session"
    echo "session-$(date +%Y%m%d-%H%M%S)" > "$SESSION_FILE"
    log_success "Created session ID file: $SESSION_FILE"
}

# Test the installation
test_installation() {
    log_info "Testing ccstatus installation..."
    
    # Test with sample data
    TEST_JSON='{"model":{"display_name":"Sonnet 4"},"workspace":{"current_dir":"'"$(pwd)"'"},"inputTokens":1500,"outputTokens":750,"contextUsage":{"tokens":25000}}'
    
    if echo "$TEST_JSON" | ccstatus > /dev/null 2>&1; then
        log_success "ccstatus is working correctly!"
        
        log_info "Sample output:"
        echo "$TEST_JSON" | ccstatus
    else
        log_error "ccstatus test failed"
        exit 1
    fi
}

# Main installation function
main() {
    log_info "Starting ccstatus installation..."
    log_info "================================"
    
    # Check permissions
    check_permissions
    
    # Install dependencies
    log_info "Installing dependencies..."
    install_go
    install_nodejs
    install_ccusage
    
    # Build and install ccstatus
    log_info "Building and installing ccstatus..."
    build_ccstatus
    install_binary
    
    # Setup Claude configuration
    log_info "Configuring Claude Code..."
    setup_claude_config
    update_statusline_config
    setup_tracking_files
    
    # Test installation
    test_installation
    
    log_info "================================"
    log_success "Installation completed successfully!"
    echo
    log_info "What's installed:"
    echo "  • ccstatus binary -> /usr/local/bin/ccstatus"
    echo "  • ccusage CLI tool -> $(which ccusage 2>/dev/null || echo 'npm global')"
    echo "  • Claude config -> ~/.claude/settings.json"
    echo "  • Tracking files -> ~/.claude/"
    echo
    log_info "Your Claude Code status line is now configured to use ccstatus!"
    echo
    log_info "Available themes (set via CCSTATUS_THEME environment variable):"
    echo "  • powerline (default) - Full Powerline theme with backgrounds"
    echo "  • minimal - Clean theme with simple separators"
    echo "  • gruvbox - Beautiful truecolor theme"
    echo
    log_info "Example: export CCSTATUS_THEME=gruvbox"
    echo
    log_info "Restart Claude Code to see your enhanced status line!"
}

# Run main function
main "$@"