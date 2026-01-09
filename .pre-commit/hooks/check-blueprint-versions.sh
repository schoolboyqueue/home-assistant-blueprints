#!/bin/bash
# Check version consistency in blueprints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

REPO_ROOT="$(git rev-parse --show-toplevel)"

# Get unique blueprint directories from changed files
BLUEPRINT_DIRS=()
for file in "$@"; do
    if [[ "$file" =~ ^blueprints/[^/]+/ ]]; then
        dir=$(dirname "$file")
        # Add to array if not already present
        if [[ ! " ${BLUEPRINT_DIRS[@]} " =~ " ${dir} " ]]; then
            BLUEPRINT_DIRS+=("$dir")
        fi
    fi
done

if [[ ${#BLUEPRINT_DIRS[@]} -eq 0 ]]; then
    exit 0
fi

VERSION_ERROR=0

for blueprint_dir in "${BLUEPRINT_DIRS[@]}"; do
    # Find the blueprint YAML file (not in shared/)
    blueprint=$(find "$REPO_ROOT/$blueprint_dir" -maxdepth 1 -name "*.yaml" -type f | head -1)
    changelog="$REPO_ROOT/$blueprint_dir/CHANGELOG.md"

    if [[ -z "$blueprint" ]]; then
        continue
    fi

    # Extract version from blueprint name field
    blueprint_name_version=$(grep -E '^\s*name:.*v[0-9]+\.[0-9]+\.[0-9]+' "$blueprint" | head -1 | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1 | sed 's/^v//' || true)

    # Extract version from blueprint_version variable
    blueprint_var_version=$(grep -E '^\s*blueprint_version:' "$blueprint" | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || true)

    # Extract latest version from CHANGELOG
    if [[ -f "$changelog" ]]; then
        changelog_version=$(grep -E '^\s*##\s*\[' "$changelog" | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || true)
    else
        changelog_version=""
    fi

    if [[ -n "$blueprint_name_version" && -n "$blueprint_var_version" ]]; then
        if [[ "$blueprint_name_version" != "$blueprint_var_version" ]]; then
            echo -e "${RED}Version mismatch in $blueprint:${NC}"
            echo -e "  Blueprint name version: ${YELLOW}$blueprint_name_version${NC}"
            echo -e "  blueprint_version var:  ${YELLOW}$blueprint_var_version${NC}"
            echo -e "${RED}Please ensure both versions match!${NC}"
            VERSION_ERROR=1
        fi

        if [[ -n "$changelog_version" && "$blueprint_name_version" != "$changelog_version" ]]; then
            echo -e "${YELLOW}Warning: CHANGELOG version ($changelog_version) differs from blueprint version ($blueprint_name_version) in $blueprint_dir${NC}"
            echo -e "${YELLOW}Consider updating CHANGELOG.md with the new version${NC}"
        fi
    fi
done

if [[ $VERSION_ERROR -eq 1 ]]; then
    exit 1
fi

exit 0
