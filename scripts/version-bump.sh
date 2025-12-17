#!/bin/bash
# version-bump.sh - Detect version bump type (patch/minor/major) using AI or conventional commits

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load environment variables if .env exists
if [ -f "$REPO_ROOT/.env" ]; then
    export $(grep -v '^#' "$REPO_ROOT/.env" | xargs)
fi

# Default to patch if no commits
if [ -z "$1" ]; then
    echo "patch"
    exit 0
fi

COMMITS="$1"

# Function to detect version bump using conventional commits
detect_conventional() {
    local commits="$1"
    local bump_type="patch"
    
    # Check for BREAKING changes (major)
    if echo "$commits" | grep -qiE "BREAKING|!:"; then
        echo "major"
        return 0
    fi
    
    # Check for feat: (minor)
    if echo "$commits" | grep -qiE "^feat\(|^feat:"; then
        bump_type="minor"
    fi
    
    # Check for fix: (patch, but don't downgrade)
    if echo "$commits" | grep -qiE "^fix\(|^fix:"; then
        if [ "$bump_type" = "patch" ]; then
            bump_type="patch"
        fi
    fi
    
    echo "$bump_type"
}

# Function to detect version bump using AI (OpenRouter)
detect_with_ai() {
    local commits="$1"
    local api_key="${OPENROUTER_API_KEY:-}"
    
    if [ -z "$api_key" ]; then
        return 1
    fi
    
    # Use a free model
    local model="google/gemini-flash-1.5-8b"
    
    # Prepare the prompt
    local prompt="Analyze these git commit messages and determine the semantic version bump type (patch, minor, or major).

Rules:
- MAJOR: Breaking changes, API changes, incompatible changes
- MINOR: New features, new functionality (backward compatible)
- PATCH: Bug fixes, documentation, refactoring (backward compatible)

Commit messages:
$commits

Respond with ONLY one word: patch, minor, or major"

    # Call OpenRouter API
    local response=$(curl -s -X POST https://openrouter.ai/api/v1/chat/completions \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $api_key" \
        -H "HTTP-Referer: https://github.com/imemir/gitext" \
        -H "X-Title: gitext release automation" \
        -d "{
            \"model\": \"$model\",
            \"messages\": [
                {
                    \"role\": \"user\",
                    \"content\": \"$prompt\"
                }
            ],
            \"temperature\": 0.1
        }" 2>/dev/null)
    
    if [ $? -ne 0 ] || [ -z "$response" ]; then
        return 1
    fi
    
    # Extract the response
    local result=$(echo "$response" | grep -oE '"content"[^}]*' | grep -oE '(patch|minor|major)' | head -1)
    
    if [ -n "$result" ]; then
        echo "$result"
        return 0
    fi
    
    return 1
}

# Try AI first, fallback to conventional commits
if detect_with_ai "$COMMITS" 2>/dev/null; then
    exit 0
else
    detect_conventional "$COMMITS"
fi
