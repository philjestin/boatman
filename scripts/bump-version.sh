#!/bin/bash
set -e

# Bump version script for CLI or Desktop
# Usage: ./scripts/bump-version.sh [cli|desktop] [major|minor|patch]

COMPONENT=$1
BUMP_TYPE=$2

if [ -z "$COMPONENT" ] || [ -z "$BUMP_TYPE" ]; then
    echo "Usage: $0 [cli|desktop|platform] [major|minor|patch]"
    echo ""
    echo "Examples:"
    echo "  $0 cli patch      # 1.2.3 -> 1.2.4"
    echo "  $0 cli minor      # 1.2.3 -> 1.3.0"
    echo "  $0 cli major      # 1.2.3 -> 2.0.0"
    echo "  $0 desktop minor  # 1.0.0 -> 1.1.0"
    echo "  $0 platform patch # 1.0.0 -> 1.0.1"
    exit 1
fi

if [ "$COMPONENT" != "cli" ] && [ "$COMPONENT" != "desktop" ] && [ "$COMPONENT" != "platform" ]; then
    echo "Error: Component must be 'cli', 'desktop', or 'platform'"
    exit 1
fi

if [ "$BUMP_TYPE" != "major" ] && [ "$BUMP_TYPE" != "minor" ] && [ "$BUMP_TYPE" != "patch" ]; then
    echo "Error: Bump type must be 'major', 'minor', or 'patch'"
    exit 1
fi

# Get current version
if [ "$COMPONENT" = "cli" ]; then
    if [ ! -f "cli/VERSION" ]; then
        echo "v0.0.0" > cli/VERSION
    fi
    CURRENT_VERSION=$(cat cli/VERSION | sed 's/v//')
elif [ "$COMPONENT" = "platform" ]; then
    if [ ! -f "platform/VERSION" ]; then
        echo "v0.1.0" > platform/VERSION
    fi
    CURRENT_VERSION=$(cat platform/VERSION | sed 's/v//')
else
    # Desktop version from wails.json
    CURRENT_VERSION=$(grep '"version"' desktop/wails.json | sed 's/.*"version": "\(.*\)".*/\1/')
fi

echo "Current $COMPONENT version: v$CURRENT_VERSION"

# Parse version
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Bump version
case $BUMP_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
esac

NEW_VERSION="$MAJOR.$MINOR.$PATCH"
echo "New $COMPONENT version: v$NEW_VERSION"

# Update version files
if [ "$COMPONENT" = "cli" ]; then
    # Update VERSION file
    echo "v$NEW_VERSION" > cli/VERSION

    # Update version.go if it exists
    if [ -f "cli/version.go" ]; then
        sed -i.bak "s/var Version = \".*\"/var Version = \"$NEW_VERSION\"/" cli/version.go
        rm cli/version.go.bak
    fi

    echo "✓ Updated cli/VERSION"
elif [ "$COMPONENT" = "platform" ]; then
    # Update VERSION file
    echo "v$NEW_VERSION" > platform/VERSION

    echo "✓ Updated platform/VERSION"
else
    # Update wails.json
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/\"version\": \".*\"/\"version\": \"$NEW_VERSION\"/" desktop/wails.json
    else
        sed -i "s/\"version\": \".*\"/\"version\": \"$NEW_VERSION\"/" desktop/wails.json
    fi

    echo "✓ Updated desktop/wails.json"
fi

# Show changelog reminder
echo ""
echo "Next steps:"
echo "  1. Update changelog: vim $COMPONENT/CHANGELOG.md"
echo "  2. Commit changes:   git add $COMPONENT && git commit -m \"$COMPONENT: Release v$NEW_VERSION\""
echo "  3. Tag release:      git tag $COMPONENT/v$NEW_VERSION"
echo "  4. Push:             git push origin main --tags"
