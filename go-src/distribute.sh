#!/bin/bash
set -e

# Coolify Go Distribution Script
# Provides multiple installation methods for different user scenarios

VERSION=${1:-"latest"}
REGISTRY="shrtso.azurecr.io/coolify-go"
GITHUB_REPO="https://github.com/s3ig8u/coolify-go-version.git"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Function to show banner
show_banner() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    🚀 Coolify Go Distribution                ║"
    echo "║                                                              ║"
    echo "║  Choose your preferred installation method                   ║"
    echo "║  Fast, secure, and production-ready deployment               ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Function to show method
show_method() {
    echo -e "${PURPLE}$1${NC}"
    echo -e "${CYAN}$2${NC}"
    echo ""
}

# Function to show command
show_command() {
    echo -e "${YELLOW}$1${NC}"
    echo ""
}

# Function to show description
show_description() {
    echo -e "${GREEN}$1${NC}"
    echo ""
}

# Function to show separator
show_separator() {
    echo -e "${BLUE}──────────────────────────────────────────────────────────────────${NC}"
    echo ""
}

# Main distribution menu
main_menu() {
    show_banner
    
    echo -e "${BLUE}Available Installation Methods:${NC}"
    echo ""
    
    show_method "1️⃣  One-Line Install (Recommended)" "Fastest installation with automatic setup"
    show_method "2️⃣  Docker Run" "Quick container deployment"
    show_method "3️⃣  Docker Compose" "Full stack with database and cache"
    show_method "4️⃣  Binary Download" "Direct binary installation"
    show_method "5️⃣  Source Build" "Build from source code"
    show_method "6️⃣  Package Manager" "System package installation"
    show_method "7️⃣  Cloud Provider" "Deploy to cloud platforms"
    show_method "8️⃣  Development Setup" "Local development environment"
    
    show_separator
    
    echo -e "${YELLOW}Select installation method (1-8):${NC}"
    read -p "Enter your choice: " choice
    
    case $choice in
        1) one_line_install ;;
        2) docker_run ;;
        3) docker_compose ;;
        4) binary_download ;;
        5) source_build ;;
        6) package_manager ;;
        7) cloud_provider ;;
        8) development_setup ;;
        *) echo -e "${RED}Invalid choice. Please select 1-8.${NC}" && main_menu ;;
    esac
}

# Method 1: One-Line Install
one_line_install() {
    show_separator
    show_description "🚀 One-Line Installation (Recommended)"
    echo "This method provides the best user experience with:"
    echo "• Automatic platform detection"
    echo "• Docker installation and configuration"
    echo "• Secure password generation"
    echo "• Full stack deployment (app + database + cache)"
    echo "• Health monitoring and validation"
    echo "• Post-installation guidance"
    echo ""
    
    show_command "For VPS/Server installation:"
    echo -e "${YELLOW}curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash${NC}"
    echo ""
    
    show_command "With custom options:"
    echo -e "${YELLOW}curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash -s -- --port 3000 --data-dir /opt/coolify${NC}"
    echo ""
    
    show_command "Available options:"
    echo "  --port PORT        Port to run on (default: 8080)"
    echo "  --data-dir DIR     Data directory (default: /data/coolify-go)"
    echo "  --skip-docker      Skip Docker installation"
    echo "  --skip-registry    Build from source instead of registry"
    echo "  --quiet            Quiet mode"
    echo ""
    
    show_command "After installation:"
    echo "• Access: http://localhost:8080"
    echo "• Health check: http://localhost:8080/health"
    echo "• Logs: docker logs coolify-go"
    echo ""
}

# Method 2: Docker Run
docker_run() {
    show_separator
    show_description "🐳 Docker Run Installation"
    echo "Quick container deployment for testing and development:"
    echo "• Minimal setup required"
    echo "• No database included (uses in-memory storage)"
    echo "• Perfect for testing and evaluation"
    echo ""
    
    show_command "Basic Docker run:"
    echo -e "${YELLOW}docker run -d \\${NC}"
    echo -e "${YELLOW}  --name coolify-go \\${NC}"
    echo -e "${YELLOW}  --restart unless-stopped \\${NC}"
    echo -e "${YELLOW}  -p 8080:8080 \\${NC}"
    echo -e "${YELLOW}  $REGISTRY:latest${NC}"
    echo ""
    
    show_command "With custom port:"
    echo -e "${YELLOW}docker run -d \\${NC}"
    echo -e "${YELLOW}  --name coolify-go \\${NC}"
    echo -e "${YELLOW}  --restart unless-stopped \\${NC}"
    echo -e "${YELLOW}  -p 3000:8080 \\${NC}"
    echo -e "${YELLOW}  $REGISTRY:latest${NC}"
    echo ""
    
    show_command "With persistent data:"
    echo -e "${YELLOW}docker run -d \\${NC}"
    echo -e "${YELLOW}  --name coolify-go \\${NC}"
    echo -e "${YELLOW}  --restart unless-stopped \\${NC}"
    echo -e "${YELLOW}  -p 8080:8080 \\${NC}"
    echo -e "${YELLOW}  -v /data/coolify-go:/data \\${NC}"
    echo -e "${YELLOW}  $REGISTRY:latest${NC}"
    echo ""
    
    show_command "Verify installation:"
    echo -e "${YELLOW}curl http://localhost:8080/health${NC}"
    echo ""
}

# Method 3: Docker Compose
docker_compose() {
    show_separator
    show_description "📦 Docker Compose Installation"
    echo "Full stack deployment with database and cache:"
    echo "• Complete production setup"
    echo "• PostgreSQL database"
    echo "• Redis cache"
    echo "• Persistent data storage"
    echo "• Health checks and monitoring"
    echo ""
    
    show_command "Download compose file:"
    echo -e "${YELLOW}curl -o docker-compose.yml https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/docker-compose.yml${NC}"
    echo ""
    
    show_command "Start services:"
    echo -e "${YELLOW}docker-compose up -d${NC}"
    echo ""
    
    show_command "Useful commands:"
    echo "• Start: docker-compose up -d"
    echo "• Stop: docker-compose down"
    echo "• Logs: docker-compose logs -f"
    echo "• Update: docker-compose pull && docker-compose up -d"
    echo ""
}

# Method 4: Binary Download
binary_download() {
    show_separator
    show_description "📥 Binary Download Installation"
    echo "Direct binary installation for maximum control:"
    echo "• No Docker required"
    echo "• Cross-platform support"
    echo "• System service integration"
    echo "• Custom configuration"
    echo ""
    
    show_command "Linux AMD64:"
    echo -e "${YELLOW}curl -L -o coolify-go https://github.com/s3ig8u/coolify-go-version/releases/latest/download/coolify-go-linux-amd64${NC}"
    echo -e "${YELLOW}chmod +x coolify-go${NC}"
    echo -e "${YELLOW}./coolify-go${NC}"
    echo ""
    
    show_command "macOS ARM64 (Apple Silicon):"
    echo -e "${YELLOW}curl -L -o coolify-go https://github.com/s3ig8u/coolify-go-version/releases/latest/download/coolify-go-darwin-arm64${NC}"
    echo -e "${YELLOW}chmod +x coolify-go${NC}"
    echo -e "${YELLOW}./coolify-go${NC}"
    echo ""
    
    show_command "Windows:"
    echo -e "${YELLOW}curl -L -o coolify-go.exe https://github.com/s3ig8u/coolify-go-version/releases/latest/download/coolify-go-windows-amd64.exe${NC}"
    echo -e "${YELLOW}./coolify-go.exe${NC}"
    echo ""
}

# Method 5: Source Build
source_build() {
    show_separator
    show_description "🔨 Source Build Installation"
    echo "Build from source for custom modifications:"
    echo "• Full control over the build process"
    echo "• Custom modifications and patches"
    echo "• Development and testing"
    echo "• Offline installation capability"
    echo ""
    
    show_command "Clone repository:"
    echo -e "${YELLOW}git clone $GITHUB_REPO${NC}"
    echo -e "${YELLOW}cd coolify-go-version/go-src${NC}"
    echo ""
    
    show_command "Build binary:"
    echo -e "${YELLOW}go build -o coolify-go .${NC}"
    echo ""
    
    show_command "Build for specific platform:"
    echo -e "${YELLOW}GOOS=linux GOARCH=amd64 go build -o coolify-go-linux-amd64 .${NC}"
    echo -e "${YELLOW}GOOS=darwin GOARCH=arm64 go build -o coolify-go-darwin-arm64 .${NC}"
    echo ""
    
    show_command "Build Docker image:"
    echo -e "${YELLOW}docker build -t coolify-go:latest .${NC}"
    echo ""
    
    show_command "Run from source:"
    echo -e "${YELLOW}go run .${NC}"
    echo ""
}

# Method 6: Package Manager
package_manager() {
    show_separator
    show_description "📦 Package Manager Installation"
    echo "System package installation (Future):"
    echo "• Native system integration"
    echo "• Automatic updates"
    echo "• Dependency management"
    echo "• Standard installation paths"
    echo ""
    
    show_command "Homebrew (macOS):"
    echo -e "${YELLOW}brew install coolify/tap/coolify-go${NC}"
    echo ""
    
    show_command "Chocolatey (Windows):"
    echo -e "${YELLOW}choco install coolify-go${NC}"
    echo ""
    
    show_command "APT (Ubuntu/Debian):"
    echo -e "${YELLOW}curl -fsSL https://packages.coolify.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/coolify-archive-keyring.gpg${NC}"
    echo -e "${YELLOW}echo 'deb [signed-by=/usr/share/keyrings/coolify-archive-keyring.gpg] https://packages.coolify.io/apt stable main' | sudo tee /etc/apt/sources.list.d/coolify.list${NC}"
    echo -e "${YELLOW}sudo apt update && sudo apt install coolify-go${NC}"
    echo ""
    
    show_warning "Note: Package manager support is planned for future releases"
    echo ""
}

# Method 7: Cloud Provider
cloud_provider() {
    show_separator
    show_description "☁️ Cloud Provider Deployment"
    echo "Deploy to cloud platforms:"
    echo "• AWS, Google Cloud, Azure"
    echo "• Kubernetes deployment"
    echo "• Serverless options"
    echo "• Managed database integration"
    echo ""
    
    show_command "Docker Compose on VPS:"
    echo -e "${YELLOW}# Use the one-line install script on any VPS${NC}"
    echo -e "${YELLOW}curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash${NC}"
    echo ""
    
    show_command "Kubernetes deployment:"
    echo -e "${YELLOW}# Create k8s-deployment.yaml with the provided template${NC}"
    echo -e "${YELLOW}kubectl apply -f k8s-deployment.yaml${NC}"
    echo ""
    
    show_warning "Note: Cloud-specific deployments may require additional configuration"
    echo ""
}

# Method 8: Development Setup
development_setup() {
    show_separator
    show_description "🛠️ Development Setup"
    echo "Local development environment:"
    echo "• Hot reload for development"
    echo "• Debug configuration"
    echo "• Testing environment"
    echo "• Local database setup"
    echo ""
    
    show_command "Prerequisites:"
    echo -e "${YELLOW}# Install Go 1.21+${NC}"
    echo -e "${YELLOW}# Install Docker${NC}"
    echo -e "${YELLOW}# Install Git${NC}"
    echo ""
    
    show_command "Clone and setup:"
    echo -e "${YELLOW}git clone $GITHUB_REPO${NC}"
    echo -e "${YELLOW}cd coolify-go-version/go-src${NC}"
    echo -e "${YELLOW}go mod download${NC}"
    echo ""
    
    show_command "Development with live reload:"
    echo -e "${YELLOW}go install github.com/cosmtrek/air@latest${NC}"
    echo -e "${YELLOW}air${NC}"
    echo ""
    
    show_command "Run tests:"
    echo -e "${YELLOW}go test ./...${NC}"
    echo -e "${YELLOW}go test -v -race ./...${NC}"
    echo ""
    
    show_command "Build and run:"
    echo -e "${YELLOW}go build -o coolify-go .${NC}"
    echo -e "${YELLOW}./coolify-go${NC}"
    echo ""
}

# Show help
show_help() {
    echo "Usage: $0 [VERSION] [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  VERSION           Version to distribute (default: latest)"
    echo "  --help            Show this help message"
    echo "  --method METHOD   Direct method selection (1-8)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Interactive menu"
    echo "  $0 v1.4.0            # Specific version"
    echo "  $0 --method 1        # Direct one-line install info"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help)
            show_help
            exit 0
            ;;
        --method)
            METHOD="$2"
            shift 2
            ;;
        -*)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

# Direct method selection
if [ -n "$METHOD" ]; then
    case $METHOD in
        1) one_line_install ;;
        2) docker_run ;;
        3) docker_compose ;;
        4) binary_download ;;
        5) source_build ;;
        6) package_manager ;;
        7) cloud_provider ;;
        8) development_setup ;;
        *) echo -e "${RED}Invalid method. Please select 1-8.${NC}" && exit 1 ;;
    esac
    exit 0
fi

# Show interactive menu
main_menu
