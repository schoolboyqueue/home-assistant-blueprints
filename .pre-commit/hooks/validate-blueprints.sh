#!/bin/bash
# Blueprint validation hook for pre-commit

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

REPO_ROOT="$(git rev-parse --show-toplevel)"
VALIDATOR="$REPO_ROOT/scripts/validate-blueprint-go/build/validate-blueprint"

# Check if validator exists
if [[ ! -f "$VALIDATOR" ]]; then
    echo -e "${YELLOW}Warning: Blueprint validator not found at $VALIDATOR${NC}"
    echo -e "${YELLOW}Run 'npm run go:build' to build the validator${NC}"
    echo -e "${YELLOW}Skipping blueprint validation${NC}"
    exit 0
fi

# Validate each blueprint file passed as argument
VALIDATION_FAILED=0

for blueprint in "$@"; do
    # Skip if file doesn't exist (might be deleted)
    if [[ ! -f "$blueprint" ]]; then
        continue
    fi

    echo "Validating $blueprint..."

    if ! "$VALIDATOR" "$blueprint" 2>&1; then
        VALIDATION_FAILED=1
        echo -e "${RED}FAILED${NC}: $blueprint"
    else
        echo -e "${GREEN}OK${NC}: $blueprint"
    fi
done

if [[ $VALIDATION_FAILED -eq 1 ]]; then
    echo ""
    echo -e "${RED}Blueprint validation failed!${NC}"
    echo -e "Please fix the issues above before committing."
    echo -e "Run manually: ${YELLOW}./scripts/validate-blueprint-go/build/validate-blueprint <blueprint.yaml>${NC}"
    exit 1
fi

exit 0
