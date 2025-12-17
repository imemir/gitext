#!/bin/bash
# release.sh - Main release automation script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Load environment variables if .env exists
if [ -f "$REPO_ROOT/.env" ]; then
    export $(grep -v '^#' "$REPO_ROOT/.env" | xargs)
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="gitext"
BUILD_DIR="build"
REMOTE="${GIT_REMOTE:-origin}"
BRANCH="${GIT_BRANCH:-master}"
DRY_RUN="${DRY_RUN:-false}"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    # Check if we're on the correct branch
    current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$current_branch" != "$BRANCH" ]; then
        log_warn "Current branch is '$current_branch', expected '$BRANCH'"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Check if working tree is clean
    if ! git diff-index --quiet HEAD --; then
        log_error "Working tree is not clean. Please commit or stash changes."
        exit 1
    fi
    
    # Check if remote exists
    if ! git remote get-url "$REMOTE" > /dev/null 2>&1; then
        log_error "Remote '$REMOTE' not found"
        exit 1
    fi
    
    # Check required commands
    for cmd in git go make curl; do
        if ! command -v $cmd > /dev/null 2>&1; then
            log_error "Required command '$cmd' not found"
            exit 1
        fi
    done
    
    log_info "Prerequisites check passed"
}

# Get latest tag
get_latest_tag() {
    log_info "Finding latest release tag..."
    
    # Fetch tags from remote
    git fetch "$REMOTE" --tags --quiet 2>/dev/null || true
    
    # Get the latest tag matching v* pattern
    latest_tag=$(git tag -l "v*" | sort -V | tail -1)
    
    if [ -z "$latest_tag" ]; then
        log_warn "No previous tags found. Starting from v0.1.0"
        echo "v0.1.0"
        return
    fi
    
    log_info "Latest tag: $latest_tag"
    echo "$latest_tag"
}

# Get commits since last tag
get_commits_since_tag() {
    local tag="$1"
    
    if [ "$tag" = "v0.1.0" ] && ! git rev-parse "$tag" > /dev/null 2>&1; then
        # First release - get all commits
        git log --pretty=format:"%s" --no-merges
    else
        # Get commits since tag
        git log "${tag}..HEAD" --pretty=format:"%s" --no-merges
    fi
}

# Calculate next version
calculate_next_version() {
    local current_tag="$1"
    local commits="$2"
    
    log_info "Determining version bump type..."
    
    # Extract version number (remove 'v' prefix)
    current_version="${current_tag#v}"
    
    # Get bump type using version-bump script
    bump_type=$("$SCRIPT_DIR/version-bump.sh" "$commits")
    
    log_info "Version bump type: $bump_type"
    
    # Parse version components
    IFS='.' read -ra VERSION_PARTS <<< "$current_version"
    major="${VERSION_PARTS[0]:-0}"
    minor="${VERSION_PARTS[1]:-0}"
    patch="${VERSION_PARTS[2]:-0}"
    
    # Calculate next version
    case "$bump_type" in
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
            log_error "Unknown bump type: $bump_type"
            exit 1
            ;;
    esac
    
    next_version="v${major}.${minor}.${patch}"
    log_info "Next version: $next_version"
    echo "$next_version"
}

# Generate changelog
generate_changelog() {
    local version="$1"
    local commits="$2"
    local previous_version="$3"
    
    log_info "Generating changelog..."
    
    changelog=$("$SCRIPT_DIR/changelog-generator.sh" "$version" "$commits" "$previous_version")
    echo "$changelog"
}

# Build binaries
build_binaries() {
    log_info "Building binaries for all platforms..."
    
    make clean
    make build-all
    
    log_info "Build complete. Binaries are in $BUILD_DIR/"
}

# Create git tag
create_tag() {
    local version="$1"
    local changelog="$2"
    
    log_info "Creating git tag: $version"
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "[DRY RUN] Would create tag: $version"
        log_info "[DRY RUN] Tag message:"
        echo "$changelog" | head -20
        return
    fi
    
    # Create annotated tag
    echo "$changelog" | git tag -a "$version" -F -
    
    log_info "Tag created: $version"
}

# Push tag to remote
push_tag() {
    local version="$1"
    
    log_info "Pushing tag to remote..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "[DRY RUN] Would push tag: $version to $REMOTE"
        return
    fi
    
    git push "$REMOTE" "$version"
    
    log_info "Tag pushed successfully"
}

# Main release process
main() {
    log_info "Starting release process..."
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --remote)
                REMOTE="$2"
                shift 2
                ;;
            --branch)
                BRANCH="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Check prerequisites
    check_prerequisites
    
    # Fetch latest changes
    log_info "Fetching latest changes from $REMOTE..."
    git fetch "$REMOTE" "$BRANCH" --quiet
    
    # Get latest tag
    latest_tag=$(get_latest_tag)
    
    # Get commits since last tag
    commits=$(get_commits_since_tag "$latest_tag")
    
    if [ -z "$commits" ]; then
        log_warn "No new commits since $latest_tag"
        exit 0
    fi
    
    # Calculate next version
    next_version=$(calculate_next_version "$latest_tag" "$commits")
    
    # Generate changelog
    changelog=$(generate_changelog "$next_version" "$commits" "$latest_tag")
    
    # Show summary
    log_info "Release summary:"
    echo "  Current tag: $latest_tag"
    echo "  Next version: $next_version"
    echo "  Commits: $(echo "$commits" | wc -l | tr -d ' ') commits"
    echo ""
    echo "Changelog preview:"
    echo "$changelog" | head -30
    echo ""
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "[DRY RUN] Release process would continue..."
        log_info "[DRY RUN] Would build binaries, create tag, and push"
        exit 0
    fi
    
    # Confirm release
    read -p "Create release $next_version? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Release cancelled"
        exit 0
    fi
    
    # Build binaries
    build_binaries
    
    # Create tag
    create_tag "$next_version" "$changelog"
    
    # Push tag
    push_tag "$next_version"
    
    log_info "Release $next_version created successfully!"
    log_info "GitHub Actions will automatically create the release and upload assets."
}

# Run main function
main "$@"
