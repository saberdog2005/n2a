#!/bin/bash

# Note2Anki Installation Script
# This script helps install and set up the Note2Anki CLI tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or higher."
        echo "Visit: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.21"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is too old. Please upgrade to Go $REQUIRED_VERSION or higher."
        exit 1
    fi
    
    print_success "Go $GO_VERSION detected"
    
    # Check for git (optional but recommended)
    if command -v git &> /dev/null; then
        print_success "Git detected"
    else
        print_info "Git not found. Some features may be limited."
    fi
}

# Install dependencies
install_dependencies() {
    echo ""
    echo "Installing Go dependencies..."
    
    go mod download
    
    if [ $? -eq 0 ]; then
        print_success "Dependencies installed successfully"
    else
        print_error "Failed to install dependencies"
        exit 1
    fi
}

# Build the binary
build_binary() {
    echo ""
    echo "Building Note2Anki..."
    
    # Create build directory
    mkdir -p build
    
    # Build with optimizations
    go build -ldflags="-s -w" -o build/note2anki main.go
    
    if [ $? -eq 0 ]; then
        print_success "Build successful"
        print_info "Binary location: $(pwd)/build/note2anki"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Install globally (optional)
install_global() {
    echo ""
    read -p "Do you want to install Note2Anki globally? (y/n) " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Installing globally..."
        
        # Try to install to /usr/local/bin (may require sudo)
        if [ -w "/usr/local/bin" ]; then
            cp build/note2anki /usr/local/bin/
            print_success "Installed to /usr/local/bin/note2anki"
        else
            print_info "Need sudo access to install globally"
            sudo cp build/note2anki /usr/local/bin/
            if [ $? -eq 0 ]; then
                print_success "Installed to /usr/local/bin/note2anki"
            else
                print_error "Failed to install globally"
            fi
        fi
    fi
}

# Setup configuration
setup_config() {
    echo ""
    echo "Setting up configuration..."
    
    # Check for existing config
    if [ -f "config.json" ]; then
        print_info "config.json already exists"
    else
        # Check for API key in environment
        if [ -z "$ANTHROPIC_API_KEY" ]; then
            echo ""
            print_info "Anthropic API key not found in environment"
            echo "You can set it by running:"
            echo "  export ANTHROPIC_API_KEY='your-api-key-here'"
            echo ""
            read -p "Would you like to enter your API key now? (y/n) " -n 1 -r
            echo
            
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                read -p "Enter your Anthropic API key: " API_KEY
                
                # Create config file
                cat > config.json <<EOF
{
  "api_key": "$API_KEY",
  "model": "gpt-3.5-turbo",
  "max_tokens": 2000,
  "temperature": 0.7
}
EOF
                print_success "Configuration file created"
            else
                # Create example config
                cp config.json.example config.json 2>/dev/null || cat > config.json <<EOF
{
  "api_key": "your-anthropic-api-key-here",
  "model": "claude-3-5-haiku-20241022",
  "max_tokens": 2000,
  "temperature": 0.7
}
EOF
                print_info "Example configuration file created. Please edit config.json with your Anthropic API key."
            fi
        else
            print_success "Anthropic API key found in environment"
        fi
    fi
}

# Create examples directory
create_examples() {
    echo ""
    echo "Creating example files..."
    
    mkdir -p examples
    
    # Create sample markdown file
    cat > examples/sample.md <<'EOF'
# Sample Biology Notes

## Cell Structure
The cell is the basic unit of life. All living organisms are composed of one or more cells.

### Cell Membrane
- Phospholipid bilayer
- Selectively permeable
- Controls what enters and exits the cell

### Nucleus
- Contains genetic material (DNA)
- Control center of the cell
- Surrounded by nuclear envelope

### Mitochondria
- Powerhouse of the cell
- Produces ATP through cellular respiration
- Has its own DNA
EOF
    
    print_success "Example files created in examples/"
}

# Run tests
run_tests() {
    echo ""
    read -p "Would you like to run a test conversion? (y/n) " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [ -f "build/note2anki" ] && [ -f "examples/sample.md" ]; then
            echo "Running test conversion..."
            ./build/note2anki -dry-run examples/sample.md test_output.txt
        else
            print_error "Could not run test. Make sure build was successful."
        fi
    fi
}

# Print usage instructions
print_usage() {
    echo ""
    echo "========================================="
    echo "Installation Complete!"
    echo "========================================="
    echo ""
    echo "Usage:"
    echo "  note2anki <input-file> <output-file>"
    echo ""
    echo "Examples:"
    echo "  note2anki notes.pdf flashcards.txt"
    echo "  note2anki -dry-run lecture.docx preview.csv"
    echo ""
    echo "For more options:"
    echo "  note2anki -help"
    echo ""
    
    if [ ! -z "$ANTHROPIC_API_KEY" ] || [ -f "config.json" ]; then
        print_success "API key configured"
    else
        print_info "Remember to set your Anthropic API key before use!"
    fi
}

# Main installation flow
main() {
    clear
    echo "========================================="
    echo "Note2Anki Installation Script"
    echo "========================================="
    echo ""
    
    check_prerequisites
    install_dependencies
    build_binary
    install_global
    setup_config
    create_examples
    run_tests
    print_usage
}

# Run main function
main