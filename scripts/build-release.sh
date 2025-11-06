#!/bin/bash
# Build release binaries for all platforms
set -e

VERSION=${1:-$(git describe --tags --always --dirty)}
OUTPUT_DIR="dist"

echo "Building podlift $VERSION"
echo ""

# Clean dist directory
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Build for each platform
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_NAME="podlift_${GOOS}_${GOARCH}"
    
    echo "Building $OUTPUT_NAME..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X github.com/ekinertac/podlift/cmd/podlift/commands.Version=$VERSION -X github.com/ekinertac/podlift/cmd/podlift/commands.Commit=$(git rev-parse HEAD)" \
        -o "${OUTPUT_DIR}/${OUTPUT_NAME}" \
        ./cmd/podlift
    
    # Create checksum
    (cd "$OUTPUT_DIR" && shasum -a 256 "$OUTPUT_NAME" > "${OUTPUT_NAME}.sha256")
    
    echo "  ✓ Built ${OUTPUT_DIR}/${OUTPUT_NAME}"
done

echo ""
echo "Release artifacts created in $OUTPUT_DIR:"
ls -lh "$OUTPUT_DIR"
echo ""
echo "To create a GitHub release:"
echo "  1. Create and push a tag: git tag v$VERSION && git push origin v$VERSION"
echo "  2. Upload files from $OUTPUT_DIR/ to the GitHub release"

