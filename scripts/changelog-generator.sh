#!/bin/bash
# changelog-generator.sh - Generate changelog using AI or conventional commits

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load environment variables if .env exists
if [ -f "$REPO_ROOT/.env" ]; then
    export $(grep -v '^#' "$REPO_ROOT/.env" | xargs)
fi

VERSION="$1"
COMMITS="$2"
PREVIOUS_VERSION="${3:-}"

if [ -z "$VERSION" ] || [ -z "$COMMITS" ]; then
    echo "Usage: $0 <version> <commits> [previous_version]"
    exit 1
fi

# Function to generate changelog using conventional commits
generate_conventional() {
    local version="$1"
    local commits="$2"
    local date=$(date +%Y-%m-%d)
    
    echo "## [$version] - $date"
    echo ""
    
    # Group commits by type
    local features=$(echo "$commits" | grep -iE "^feat\(|^feat:" || true)
    local fixes=$(echo "$commits" | grep -iE "^fix\(|^fix:" || true)
    local breaking=$(echo "$commits" | grep -iE "BREAKING|!:" || true)
    local docs=$(echo "$commits" | grep -iE "^docs\(|^docs:" || true)
    local other=$(echo "$commits" | grep -vE "^feat\(|^feat:|^fix\(|^fix:|BREAKING|!:|^docs\(|^docs:" || true)
    
    if [ -n "$breaking" ]; then
        echo "### BREAKING CHANGES"
        echo "$breaking" | sed 's/^/- /' | sed 's/^BREAKING CHANGE: /- **BREAKING CHANGE:** /'
        echo ""
    fi
    
    if [ -n "$features" ]; then
        echo "### Added"
        echo "$features" | sed 's/^feat(\([^)]*\)): /- \1: /' | sed 's/^feat: /- /' | sed 's/^/- /'
        echo ""
    fi
    
    if [ -n "$fixes" ]; then
        echo "### Fixed"
        echo "$fixes" | sed 's/^fix(\([^)]*\)): /- \1: /' | sed 's/^fix: /- /' | sed 's/^/- /'
        echo ""
    fi
    
    if [ -n "$docs" ]; then
        echo "### Documentation"
        echo "$docs" | sed 's/^docs(\([^)]*\)): /- \1: /' | sed 's/^docs: /- /' | sed 's/^/- /'
        echo ""
    fi
    
    if [ -n "$other" ]; then
        echo "### Changed"
        echo "$other" | sed 's/^/- /'
        echo ""
    fi
}

# Function to generate changelog using AI (OpenRouter)
generate_with_ai() {
    local version="$1"
    local commits="$2"
    local previous_version="$3"
    local api_key="${OPENROUTER_API_KEY:-}"
    
    if [ -z "$api_key" ]; then
        return 1
    fi
    
    # Use a free model
    local model="google/gemini-flash-1.5-8b"
    local date=$(date +%Y-%m-%d)
    
    # Prepare the prompt
    local prompt="Generate a changelog entry in Markdown format for version $version released on $date.

Previous version: ${previous_version:-N/A}

Commit messages:
$commits

Format the changelog as:
## [version] - date

### Added
- Feature 1
- Feature 2

### Fixed
- Bug fix 1

### Changed
- Change 1

### BREAKING CHANGES (if any)
- Breaking change description

Group commits logically and write clear, user-friendly descriptions. Use proper Markdown formatting."

    # Call OpenRouter API
    local response=$(curl -s -X POST https://openrouter.ai/api/v1/chat/completions \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $api_key" \
        -H "HTTP-Referer: https://github.com/gitext/gitext" \
        -H "X-Title: gitext release automation" \
        -d "{
            \"model\": \"$model\",
            \"messages\": [
                {
                    \"role\": \"user\",
                    \"content\": \"$prompt\"
                }
            ],
            \"temperature\": 0.3
        }" 2>/dev/null)
    
    if [ $? -ne 0 ] || [ -z "$response" ]; then
        return 1
    fi
    
    # Extract the content from response
    local content=$(echo "$response" | grep -oE '"content"[^}]*' | sed 's/"content":"//' | sed 's/\\n/\n/g' | sed 's/\\"/"/g')
    
    if [ -n "$content" ]; then
        echo "$content"
        return 0
    fi
    
    return 1
}

# Try AI first, fallback to conventional commits
if generate_with_ai "$VERSION" "$COMMITS" "$PREVIOUS_VERSION" 2>/dev/null; then
    exit 0
else
    generate_conventional "$VERSION" "$COMMITS"
fi
