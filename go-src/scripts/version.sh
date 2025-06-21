#!/bin/bash
set -e

# Version management script for Coolify Go
# Usage: ./scripts/version.sh [major|minor|patch|set] [version]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION_FILE="$PROJECT_ROOT/version.go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to get current version from version.go
get_current_version() {
    if [ -f "$VERSION_FILE" ]; then
        grep -o 'Version[[:space:]]*=[[:space:]]*"[^"]*"' "$VERSION_FILE" | cut -d'"' -f2
    else
        echo "v0.0.0"
    fi
}

# Function to validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
        print_error "Invalid version format: $version"
        print_info "Version must follow semantic versioning: vX.Y.Z[-prerelease][+build]"
        exit 1
    fi
}

# Function to bump version
bump_version() {
    local current_version=$1
    local bump_type=$2
    
    # Remove 'v' prefix for processing
    local version_without_v=${current_version#v}
    
    # Split version into components
    IFS='.' read -ra VERSION_PARTS <<< "$version_without_v"
    local major=${VERSION_PARTS[0]}
    local minor=${VERSION_PARTS[1]}
    local patch=${VERSION_PARTS[2]}
    
    case $bump_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            print_error "Invalid bump type: $bump_type"
            print_info "Valid types: major, minor, patch"
            exit 1
            ;;
    esac
    
    echo "v$major.$minor.$patch"
}

# Function to update version.go
update_version_file() {
    local version=$1
    local build_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local git_commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    print_info "Updating version.go with version $version"
    
    cat > "$VERSION_FILE" << EOF
package main

import "fmt"

const (
	Version   = "$version"
	BuildTime = "$build_time"
	GitCommit = "$git_commit"
)

func printVersion() {
	fmt.Printf("Coolify Go v%s (built %s, commit %s)\n", Version, BuildTime, GitCommit)
}
EOF
    
    print_success "Updated version.go"
}

# Function to create git tag
create_git_tag() {
    local version=$1
    local message=${2:-"Release $version"}
    
    if git rev-parse "$version" >/dev/null 2>&1; then
        print_warning "Tag $version already exists"
        return 0
    fi
    
    print_info "Creating git tag: $version"
    git tag -a "$version" -m "$message"
    print_success "Created git tag: $version"
}

# Function to push git tag
push_git_tag() {
    local version=$1
    
    print_info "Pushing git tag: $version"
    if git push origin "$version"; then
        print_success "Pushed git tag: $version"
    else
        print_warning "Failed to push git tag (may not have remote access)"
    fi
}

# Function to show current version info
show_version_info() {
    local current_version=$(get_current_version)
    local git_commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local git_status=""
    
    if [ -n "$(git status --porcelain)" ]; then
        git_status=" (dirty)"
    fi
    
    echo "Current Version: $current_version"
    echo "Git Commit: $git_commit$git_status"
    echo "Version File: $VERSION_FILE"
}

# Main script logic
main() {
    cd "$PROJECT_ROOT"
    
    case $1 in
        "major"|"minor"|"patch")
            local current_version=$(get_current_version)
            local new_version=$(bump_version "$current_version" "$1")
            
            print_info "Bumping version from $current_version to $new_version"
            update_version_file "$new_version"
            
            # Stage version.go changes
            git add "$VERSION_FILE"
            
            print_success "Version bumped to $new_version"
            print_info "Don't forget to commit and tag:"
            echo "  git commit -m \"Bump version to $new_version\""
            echo "  git tag -a \"$new_version\" -m \"Release $new_version\""
            echo "  git push origin \"$new_version\""
            ;;
            
        "set")
            if [ -z "$2" ]; then
                print_error "Version required for 'set' command"
                print_info "Usage: $0 set v1.2.3"
                exit 1
            fi
            
            local new_version=$2
            validate_version "$new_version"
            
            print_info "Setting version to $new_version"
            update_version_file "$new_version"
            
            # Stage version.go changes
            git add "$VERSION_FILE"
            
            print_success "Version set to $new_version"
            ;;
            
        "tag")
            local current_version=$(get_current_version)
            local message=${2:-"Release $current_version"}
            
            print_info "Creating tag for version $current_version"
            create_git_tag "$current_version" "$message"
            ;;
            
        "push-tag")
            local current_version=$(get_current_version)
            push_git_tag "$current_version"
            ;;
            
        "release")
            local current_version=$(get_current_version)
            local message=${2:-"Release $current_version"}
            
            print_info "Creating release for version $current_version"
            create_git_tag "$current_version" "$message"
            push_git_tag "$current_version"
            print_success "Release $current_version created and pushed"
            ;;
            
        "info"|"show"|"current")
            show_version_info
            ;;
            
        "help"|"-h"|"--help"|"")
            echo "Version Management Script for Coolify Go"
            echo ""
            echo "Usage: $0 [command] [options]"
            echo ""
            echo "Commands:"
            echo "  major              Bump major version (X.0.0)"
            echo "  minor              Bump minor version (0.X.0)"
            echo "  patch              Bump patch version (0.0.X)"
            echo "  set <version>      Set specific version (e.g., v1.2.3)"
            echo "  tag [message]      Create git tag for current version"
            echo "  push-tag           Push current version tag to remote"
            echo "  release [message]  Create and push tag for current version"
            echo "  info|show|current  Show current version information"
            echo "  help               Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 patch                    # Bump patch version"
            echo "  $0 minor                    # Bump minor version"
            echo "  $0 set v2.0.0              # Set specific version"
            echo "  $0 release \"New features\"  # Create and push release"
            echo ""
            echo "Current version: $(get_current_version)"
            ;;
            
        *)
            print_error "Unknown command: $1"
            print_info "Run '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 