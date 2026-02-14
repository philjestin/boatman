#!/bin/bash
set -e

# Interactive release script
# Guides you through the release process

echo "🚢 Boatman Ecosystem Release Script"
echo "====================================="
echo ""

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "⚠️  Working directory is not clean. Commit or stash changes first."
    git status --short
    exit 1
fi

# Ask which component to release
echo "Which component do you want to release?"
echo "  1) CLI only"
echo "  2) Desktop only"
echo "  3) Both (coordinated release)"
read -p "Choice (1-3): " CHOICE

case $CHOICE in
    1) COMPONENTS="cli" ;;
    2) COMPONENTS="desktop" ;;
    3) COMPONENTS="cli desktop" ;;
    *) echo "Invalid choice"; exit 1 ;;
esac

# For each component, ask for bump type
declare -A VERSIONS
for COMPONENT in $COMPONENTS; do
    echo ""
    echo "=== $COMPONENT ==="

    # Show current version
    if [ "$COMPONENT" = "cli" ]; then
        CURRENT=$(cat cli/VERSION | sed 's/v//')
    else
        CURRENT=$(grep '"version"' desktop/wails.json | sed 's/.*"version": "\(.*\)".*/\1/')
    fi
    echo "Current version: v$CURRENT"

    # Ask for bump type
    echo "What type of release?"
    echo "  1) Patch (bug fixes) - ${CURRENT} -> ${CURRENT%.*}.$((${CURRENT##*.}+1))"
    IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"
    echo "  2) Minor (new features) - ${CURRENT} -> ${MAJOR}.$((MINOR+1)).0"
    echo "  3) Major (breaking changes) - ${CURRENT} -> $((MAJOR+1)).0.0"
    echo "  4) Custom version"
    read -p "Choice (1-4): " BUMP_CHOICE

    case $BUMP_CHOICE in
        1) BUMP_TYPE="patch" ;;
        2) BUMP_TYPE="minor" ;;
        3) BUMP_TYPE="major" ;;
        4)
            read -p "Enter new version (without 'v'): " NEW_VERSION
            BUMP_TYPE="custom"
            ;;
        *) echo "Invalid choice"; exit 1 ;;
    esac

    # Bump version
    if [ "$BUMP_TYPE" = "custom" ]; then
        if [ "$COMPONENT" = "cli" ]; then
            echo "v$NEW_VERSION" > cli/VERSION
        else
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s/\"version\": \".*\"/\"version\": \"$NEW_VERSION\"/" desktop/wails.json
            else
                sed -i "s/\"version\": \".*\"/\"version\": \"$NEW_VERSION\"/" desktop/wails.json
            fi
        fi
        VERSIONS[$COMPONENT]=$NEW_VERSION
    else
        ./scripts/bump-version.sh $COMPONENT $BUMP_TYPE > /dev/null
        if [ "$COMPONENT" = "cli" ]; then
            VERSIONS[$COMPONENT]=$(cat cli/VERSION | sed 's/v//')
        else
            VERSIONS[$COMPONENT]=$(grep '"version"' desktop/wails.json | sed 's/.*"version": "\(.*\)".*/\1/')
        fi
    fi

    echo "✓ Version bumped to v${VERSIONS[$COMPONENT]}"
done

echo ""
echo "=== Summary ==="
for COMPONENT in $COMPONENTS; do
    echo "  $COMPONENT: v${VERSIONS[$COMPONENT]}"
done
echo ""

# Ask for confirmation to edit changelogs
for COMPONENT in $COMPONENTS; do
    read -p "Edit $COMPONENT/CHANGELOG.md? (y/n): " EDIT
    if [ "$EDIT" = "y" ]; then
        ${EDITOR:-vim} $COMPONENT/CHANGELOG.md
    fi
done

echo ""
echo "=== Release checklist ==="
for COMPONENT in $COMPONENTS; do
    echo ""
    echo "$COMPONENT v${VERSIONS[$COMPONENT]}:"
    echo "  ✓ Version bumped"
    echo "  ? Changelog updated"
    echo "  ? Tests passing"
    echo "  ? Documentation updated"
done
echo ""

# Run tests
read -p "Run tests before releasing? (y/n): " RUN_TESTS
if [ "$RUN_TESTS" = "y" ]; then
    echo "Running tests..."
    for COMPONENT in $COMPONENTS; do
        echo "  Testing $COMPONENT..."
        (cd $COMPONENT && go test ./... || exit 1)
    done
    echo "✓ All tests passed"
fi

echo ""
echo "=== Commit and Tag ==="

# Prepare commit message
if [ "$COMPONENTS" = "cli desktop" ]; then
    COMMIT_MSG="Release: CLI v${VERSIONS[cli]}, Desktop v${VERSIONS[desktop]}"
else
    COMPONENT=${COMPONENTS}
    COMMIT_MSG="$COMPONENT: Release v${VERSIONS[$COMPONENT]}"
fi

echo "Commit message: $COMMIT_MSG"
read -p "Proceed with commit and tag? (y/n): " PROCEED

if [ "$PROCEED" != "y" ]; then
    echo "Aborting. You can manually commit with:"
    for COMPONENT in $COMPONENTS; do
        echo "  git add $COMPONENT"
    done
    echo "  git commit -m \"$COMMIT_MSG\""
    for COMPONENT in $COMPONENTS; do
        echo "  git tag $COMPONENT/v${VERSIONS[$COMPONENT]}"
    done
    exit 0
fi

# Commit
git add ${COMPONENTS}
git commit -m "$COMMIT_MSG"

# Tag
for COMPONENT in $COMPONENTS; do
    TAG="${COMPONENT}/v${VERSIONS[$COMPONENT]}"
    git tag $TAG
    echo "✓ Tagged $TAG"
done

echo ""
echo "=== Push ==="
echo "Ready to push to remote. This will trigger release workflows."
read -p "Push to origin? (y/n): " PUSH

if [ "$PUSH" = "y" ]; then
    git push origin main
    git push origin --tags
    echo ""
    echo "✅ Release pushed!"
    echo ""
    echo "GitHub Actions will now:"
    for COMPONENT in $COMPONENTS; do
        if [ "$COMPONENT" = "cli" ]; then
            echo "  • Build CLI binaries for all platforms"
            echo "  • Create GitHub release for cli/v${VERSIONS[$COMPONENT]}"
        else
            echo "  • Build desktop apps for macOS, Linux, Windows"
            echo "  • Create GitHub release for desktop/v${VERSIONS[$COMPONENT]}"
        fi
    done
    echo ""
    echo "View releases at: https://github.com/YOUR_ORG/boatman-ecosystem/releases"
else
    echo ""
    echo "Not pushed. You can push manually with:"
    echo "  git push origin main --tags"
fi

echo ""
echo "🎉 Release process complete!"
