#!/bin/bash
# Check that README and CHANGELOG exist for each blueprint

set -e

# Colors for output
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

REPO_ROOT="$(git rev-parse --show-toplevel)"

# Get unique blueprint directories from changed files
BLUEPRINT_DIRS=()
for file in "$@"; do
    if [[ "$file" =~ ^blueprints/[^/]+/ && ! "$file" =~ ^blueprints/shared/ ]]; then
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

for blueprint_dir in "${BLUEPRINT_DIRS[@]}"; do
    if [[ ! -f "$REPO_ROOT/$blueprint_dir/README.md" ]]; then
        echo -e "${YELLOW}Warning: Missing README.md in $blueprint_dir${NC}"
    fi

    if [[ ! -f "$REPO_ROOT/$blueprint_dir/CHANGELOG.md" ]]; then
        echo -e "${YELLOW}Warning: Missing CHANGELOG.md in $blueprint_dir${NC}"
    fi
done

exit 0
