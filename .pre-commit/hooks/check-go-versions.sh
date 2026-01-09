#!/bin/bash
# Check Go tool versions and changelogs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

REPO_ROOT="$(git rev-parse --show-toplevel)"

# Determine which tools need checking based on changed files
CHECK_VALIDATE=false
CHECK_CLIENT=false

for file in "$@"; do
    if [[ "$file" =~ ^scripts/validate-blueprint-go/ ]]; then
        CHECK_VALIDATE=true
    fi
    if [[ "$file" =~ ^scripts/ha-ws-client-go/ ]]; then
        CHECK_CLIENT=true
    fi
done

GO_TOOLS=()
[[ "$CHECK_VALIDATE" == true ]] && GO_TOOLS+=("validate-blueprint-go")
[[ "$CHECK_CLIENT" == true ]] && GO_TOOLS+=("ha-ws-client-go")

if [[ ${#GO_TOOLS[@]} -eq 0 ]]; then
    exit 0
fi

GO_VERSIONS=()
VERSION_ERROR=0

for tool in "${GO_TOOLS[@]}"; do
    TOOL_DIR="$REPO_ROOT/scripts/$tool"
    MAKEFILE="$TOOL_DIR/Makefile"
    CHANGELOG="$TOOL_DIR/CHANGELOG.md"

    # Check if CHANGELOG.md exists
    if [[ ! -f "$CHANGELOG" ]]; then
        echo -e "${RED}Error: Missing CHANGELOG.md in $tool${NC}"
        echo -e "  ${YELLOW}Please create $CHANGELOG following Keep a Changelog format${NC}"
        VERSION_ERROR=1
        continue
    fi

    # Extract VERSION from Makefile (default value)
    if [[ -f "$MAKEFILE" ]]; then
        MAKEFILE_VERSION=$(grep -E '^VERSION\?=' "$MAKEFILE" | head -1 | sed 's/VERSION?=//' || true)
    else
        MAKEFILE_VERSION=""
    fi

    # Extract latest version from CHANGELOG (first ## [X.Y.Z] entry)
    CHANGELOG_VERSION=$(grep -E '^\s*##\s*\[' "$CHANGELOG" | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || true)

    if [[ -n "$MAKEFILE_VERSION" && -n "$CHANGELOG_VERSION" ]]; then
        if [[ "$MAKEFILE_VERSION" != "$CHANGELOG_VERSION" ]]; then
            echo -e "${RED}Version mismatch in $tool:${NC}"
            echo -e "  Makefile VERSION:  ${YELLOW}$MAKEFILE_VERSION${NC}"
            echo -e "  CHANGELOG version: ${YELLOW}$CHANGELOG_VERSION${NC}"
            echo -e "${YELLOW}Please ensure Makefile VERSION matches the latest CHANGELOG entry${NC}"
            VERSION_ERROR=1
        else
            echo -e "${GREEN}$tool: v$MAKEFILE_VERSION${NC} ✓"
            GO_VERSIONS+=("$MAKEFILE_VERSION")
        fi
    elif [[ -z "$CHANGELOG_VERSION" ]]; then
        echo -e "${YELLOW}Warning: Could not extract version from $tool/CHANGELOG.md${NC}"
        echo -e "  ${YELLOW}Ensure CHANGELOG.md has entries like: ## [1.0.0] - YYYY-MM-DD${NC}"
    fi
done

# Check that both tools have the same version (synchronized releases)
if [[ ${#GO_VERSIONS[@]} -eq 2 ]]; then
    if [[ "${GO_VERSIONS[0]}" != "${GO_VERSIONS[1]}" ]]; then
        echo -e "${YELLOW}Warning: Go tool versions are not synchronized:${NC}"
        echo -e "  ha-ws-client-go:       v${GO_VERSIONS[0]}"
        echo -e "  validate-blueprint-go: v${GO_VERSIONS[1]}"
        echo -e "${YELLOW}Consider keeping both tools at the same version for coordinated releases${NC}"
    fi
fi

if [[ $VERSION_ERROR -eq 1 ]]; then
    echo ""
    echo -e "${RED}Go tool version/changelog check failed!${NC}"
    exit 1
fi

exit 0
