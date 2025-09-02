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

# Detect system architecture and OS
detect_system() {
    # Detect OS
    case "$(uname -s)" in
        Linux*) OS="linux" ;;
        Darwin*) OS="darwin" ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*) OS="windows" ;;
        *) log_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv6l) ARCH="armv6l" ;;
        armv7l) ARCH="armv7l" ;;
        *) log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
    
    # Set file extension
    if [ "$OS" = "windows" ]; then
        EXT=".exe"
        ARCHIVE_EXT="zip"
    else
        EXT=""
        ARCHIVE_EXT="tar.gz"
    fi
    
    log_info "Detected system: $OS/$ARCH"
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

# Get latest release version from GitHub
get_latest_version() {
    log_info "Fetching latest release version..."
    
    LATEST_VERSION=$(curl -s https://api.github.com/repos/mrdavidaylward/ccstatus/releases/latest | grep '"tag_name"' | cut -d'"' -f4)
    
    if [ -z "$LATEST_VERSION" ]; then
        log_error "Failed to fetch latest version from GitHub"
        exit 1
    fi
    
    log_success "Latest version: $LATEST_VERSION"
}

# Download and extract ccstatus binary
download_ccstatus() {
    log_info "Downloading ccstatus $LATEST_VERSION for $OS/$ARCH..."
    
    # Construct download URL
    FILENAME="ccstatus_${LATEST_VERSION}_${OS}_${ARCH}.${ARCHIVE_EXT}"
    DOWNLOAD_URL="https://github.com/mrdavidaylward/ccstatus/releases/download/$LATEST_VERSION/$FILENAME"
    
    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Download the archive
    if ! curl -L -o "$FILENAME" "$DOWNLOAD_URL"; then
        log_error "Failed to download $FILENAME"
        log_error "URL: $DOWNLOAD_URL"
        exit 1
    fi
    
    log_success "Downloaded $FILENAME"
    
    # Extract the archive
    log_info "Extracting archive..."
    if [ "$ARCHIVE_EXT" = "zip" ]; then
        unzip -q "$FILENAME"
    else
        tar -xzf "$FILENAME"
    fi
    
    # Find the extracted directory
    EXTRACT_DIR="ccstatus_${OS}_${ARCH}"
    if [ ! -d "$EXTRACT_DIR" ]; then
        log_error "Extracted directory not found: $EXTRACT_DIR"
        exit 1
    fi
    
    # Verify binary exists
    BINARY_PATH="$EXTRACT_DIR/ccstatus$EXT"
    if [ ! -f "$BINARY_PATH" ]; then
        log_error "ccstatus binary not found in extracted archive"
        exit 1
    fi
    
    # Make binary executable
    chmod +x "$BINARY_PATH"
    
    log_success "Successfully extracted ccstatus binary"
}

# Install ccstatus to system PATH
install_binary() {
    log_info "Installing ccstatus to /usr/local/bin..."
    
    # Copy binary to system location
    sudo cp "$BINARY_PATH" /usr/local/bin/ccstatus
    
    # Verify installation
    if command -v ccstatus &> /dev/null; then
        log_success "ccstatus installed to /usr/local/bin/ccstatus"
    else
        log_error "Failed to install ccstatus binary"
        exit 1
    fi
    
    # Clean up temp directory
    cd /
    rm -rf "$TEMP_DIR"
    log_info "Cleaned up temporary files"
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
    TEST_JSON='{"model":{"display_name":"Sonnet 4"},"workspace":{"current_dir":"'"$HOME"'"},"inputTokens":1500,"outputTokens":750,"contextUsage":{"tokens":25000}}'
    
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
    
    # Detect system
    detect_system
    
    # Install dependencies
    log_info "Installing dependencies..."
    install_nodejs
    install_ccusage
    
    # Download and install ccstatus
    log_info "Downloading and installing ccstatus..."
    get_latest_version
    download_ccstatus
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
    echo "  • ccstatus $LATEST_VERSION -> /usr/local/bin/ccstatus"
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