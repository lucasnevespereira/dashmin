#!/bin/bash

set -e

# Usage function
usage() {
    echo "Usage: ./release.sh <version> [--retry]"
    echo ""
    echo "Examples:"
    echo "  ./release.sh 0.1.2        # Release version 0.1.2"
    echo "  ./release.sh 1.0.0        # Release version 1.0.0"
    echo "  ./release.sh 0.1.2 --retry # Retry failed release"
    exit 1
}

# Check if version is provided
if [ $# -eq 0 ]; then
    usage
fi

NEW_VERSION=$1
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

# Validate version format (basic semver check)
if [[ ! "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ Invalid version format: $NEW_VERSION"
    echo "Please use semantic versioning format: MAJOR.MINOR.PATCH (e.g., 1.0.0)"
    usage
fi

echo "Releasing version: $NEW_VERSION"
echo ""

# Check git status first
echo "🔍 Checking git status..."
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Git working directory is not clean. Please commit or stash your changes."
    exit 1
fi

# Update version in code
echo "📝 Updating version in cmd/version.go..."
sed -i.bak "s/const Version = \".*\"/const Version = \"$NEW_VERSION\"/" cmd/version.go
rm cmd/version.go.bak

# Commit version update
echo "💾 Committing version update..."
git add cmd/version.go
git commit -m "chore: bump version to $NEW_VERSION"

# Push the version commit to main
echo "📤 Pushing version commit to main..."
git push origin main

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
echo "🚀 GitHub Actions will now create the release with proper versioning"
echo "📦 Check the release at: https://github.com/lucasnevespereira/dashmin/releases/tag/v$NEW_VERSION"