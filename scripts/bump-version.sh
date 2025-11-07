#!/bin/bash
set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get bump type (patch, minor, major, alpha, beta)
BUMP_TYPE=${1:-patch}

# Validate bump type
if [[ ! "$BUMP_TYPE" =~ ^(patch|minor|major|alpha|beta)$ ]]; then
    echo -e "${RED}Error: Invalid bump type '$BUMP_TYPE'${NC}"
    echo "Usage: $0 [patch|minor|major|alpha|beta]"
    exit 1
fi

# Get current version from git tags
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Remove 'v' prefix if present
CURRENT_VERSION=${CURRENT_VERSION#v}

# Check if current version is a pre-release
if [[ "$CURRENT_VERSION" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)-([a-z]+)\.([0-9]+)$ ]]; then
    # Current is pre-release (e.g., 1.0.3-alpha.1)
    MAJOR=${BASH_REMATCH[1]}
    MINOR=${BASH_REMATCH[2]}
    PATCH=${BASH_REMATCH[3]}
    PRERELEASE_TYPE=${BASH_REMATCH[4]}
    PRERELEASE_NUM=${BASH_REMATCH[5]}
else
    # Current is stable release (e.g., 1.0.2)
    IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
    MAJOR=${VERSION_PARTS[0]:-0}
    MINOR=${VERSION_PARTS[1]:-0}
    PATCH=${VERSION_PARTS[2]:-0}
    PRERELEASE_TYPE=""
    PRERELEASE_NUM=0
fi

# Calculate new version based on bump type
case $BUMP_TYPE in
    patch)
        NEW_MAJOR=$MAJOR
        NEW_MINOR=$MINOR
        NEW_PATCH=$((PATCH + 1))
        NEW_VERSION="v${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"
        ;;
    minor)
        NEW_MAJOR=$MAJOR
        NEW_MINOR=$((MINOR + 1))
        NEW_PATCH=0
        NEW_VERSION="v${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"
        ;;
    major)
        NEW_MAJOR=$((MAJOR + 1))
        NEW_MINOR=0
        NEW_PATCH=0
        NEW_VERSION="v${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"
        ;;
    alpha|beta)
        # For pre-releases, bump patch and add/increment pre-release number
        if [[ "$PRERELEASE_TYPE" == "$BUMP_TYPE" ]]; then
            # Same pre-release type, increment number
            NEW_MAJOR=$MAJOR
            NEW_MINOR=$MINOR
            NEW_PATCH=$PATCH
            NEW_PRERELEASE_NUM=$((PRERELEASE_NUM + 1))
        else
            # Different pre-release type or first pre-release
            NEW_MAJOR=$MAJOR
            NEW_MINOR=$MINOR
            NEW_PATCH=$((PATCH + 1))
            NEW_PRERELEASE_NUM=1
        fi
        NEW_VERSION="v${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}-${BUMP_TYPE}.${NEW_PRERELEASE_NUM}"
        ;;
esac

echo -e "${YELLOW}Current version:${NC} v${CURRENT_VERSION}"
echo -e "${YELLOW}New version:${NC}     ${NEW_VERSION}"
echo ""

# Check for uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${RED}Error: You have uncommitted changes${NC}"
    echo "Please commit or stash them before bumping version."
    exit 1
fi

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    echo -e "${RED}Error: Not on main branch (current: $CURRENT_BRANCH)${NC}"
    echo "Switch to main branch before bumping version."
    exit 1
fi

# Show what will happen
echo -e "${YELLOW}Releasing:${NC}"
echo "  • Create git tag: ${NEW_VERSION}"
echo "  • Push tag to GitHub"
echo "  • Trigger GitHub Actions release build"
echo ""

# Create annotated tag
echo ""
echo -e "${GREEN}Creating tag ${NEW_VERSION}...${NC}"
git tag -a "${NEW_VERSION}" -m "${NEW_VERSION}

Automated ${BUMP_TYPE} version bump from v${CURRENT_VERSION} to ${NEW_VERSION}"

# Push tag
echo -e "${GREEN}Pushing tag to GitHub...${NC}"
git push origin "${NEW_VERSION}"

echo ""
echo -e "${GREEN}✓ Version bumped to ${NEW_VERSION}${NC}"
echo ""
echo "GitHub Actions is now building release:"
echo "  https://github.com/ekinertac/podlift/actions"
echo ""
echo "Release will be available at:"
echo "  https://github.com/ekinertac/podlift/releases/tag/${NEW_VERSION}"
echo ""
echo "Monitor progress with:"
echo "  watch -n 5 'gh run list --limit 1'"

