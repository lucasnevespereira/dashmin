#!/bin/bash

set -e

# Usage function
usage() {
    echo "Usage: ./release.sh <patch|minor|major> [--retry]"
    echo ""
    echo "Examples:"
    echo "  ./release.sh patch        # 0.1.0 -> 0.1.1"
    echo "  ./release.sh minor        # 0.1.0 -> 0.2.0"
    echo "  ./release.sh major        # 0.1.0 -> 1.0.0"
    echo "  ./release.sh patch --retry # Retry failed release"
    exit 1
}

# Check if type is provided
if [ $# -eq 0 ]; then
    usage
fi

TYPE=$1
RETRY=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --retry)
            RETRY=true
            shift
            ;;
        *)
            ;;
    esac
done

# Validate type
if [[ ! "$TYPE" =~ ^(patch|minor|major)$ ]]; then
    echo "❌ Invalid release type: $TYPE"
    usage
fi

# Get current version from cmd/version.go
CURRENT_VERSION=$(grep 'const Version' cmd/version.go | sed 's/.*"\(.*\)".*/\1/')

if [ -z "$CURRENT_VERSION" ]; then
    echo "❌ Could not find current version in version.go"
    exit 1
fi

# Calculate new version
IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

case $TYPE in
    major)
        NEW_VERSION="$((MAJOR + 1)).0.0"
        ;;
    minor)
        NEW_VERSION="$MAJOR.$((MINOR + 1)).0"
        ;;
    patch)
        NEW_VERSION="$MAJOR.$MINOR.$((PATCH + 1))"
        ;;
esac

echo "Current version: $CURRENT_VERSION"
echo "New version: $NEW_VERSION"
echo ""

# Check git status
echo "🔍 Checking git status..."
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Git working directory is not clean. Please commit or stash your changes."
    exit 1
fi

# Handle existing tags
if git tag -l | grep -q "^v$NEW_VERSION$"; then
    echo "⚠️  Tag v$NEW_VERSION already exists locally. Deleting it..."
    git tag -d "v$NEW_VERSION"
fi

if git ls-remote --tags origin | grep -q "refs/tags/v$NEW_VERSION$"; then
    if [ "$RETRY" = true ]; then
        echo "⚠️  Tag v$NEW_VERSION exists on remote. Deleting for retry..."
        git push --delete origin "v$NEW_VERSION"
        gh release delete "v$NEW_VERSION" --yes 2>/dev/null || echo "   (No GitHub release found)"
    else
        echo "❌ Tag v$NEW_VERSION already exists on remote!"
        echo "If you want to retry this release, use: ./release.sh $TYPE --retry"
        exit 1
    fi
fi

# Create and push tag
echo "🏷️  Creating and pushing git tag..."
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
git push origin "v$NEW_VERSION"

echo "✅ Tag pushed successfully!"
echo ""
echo "🚀 GitHub Actions will now create the release and update version.go"
echo "📦 Check the release at: https://github.com/lucasnevespereira/dashmin/releases/tag/v$NEW_VERSION"