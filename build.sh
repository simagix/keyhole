#!/bin/bash
# Copyright 2020 Kuei-chun Chen. All rights reserved.
die() { echo "$*" 1>&2 ; exit 1; }

# Use git commit date if available, otherwise use current date
GIT_DATE=$(git log -1 --date=format:"%Y%m%d" --format="%ad" 2>/dev/null || date +"%Y%m%d")
VERSION="v$(cat version)-${GIT_DATE}"
REPO=$(basename "$(dirname "$(pwd)")")/$(basename "$(pwd)")
LDFLAGS="-X main.version=$VERSION -X main.repo=$REPO"
TAG="simagix/keyhole"

[[ "$(which go)" = "" ]] && die "go command not found"

gover=$(go version | cut -d' ' -f3)
if [ "$gover" \< "go1.18" ]; then
    [[ "$GOPATH" = "" ]] && die "GOPATH not set"
    [[ "${GOPATH}/src/github.com/$REPO" != "$(pwd)" ]] && die "building keyhole should be under ${GOPATH}/src/github.com/$REPO"
fi

if [ ! -f go.sum ]; then
    go mod tidy
fi

print_usage() {
  echo "Usage: $0 [command]"
  echo ""
  echo "Commands:"
  echo "  (none)        Build binary for current platform to dist/"
  echo "  docker        Build Docker image for current platform (local)"
  echo "  push          Build and push multi-arch Docker image (amd64 + arm64)"
  echo "  binaries      Build binaries for all platforms (linux/mac/win, amd64/arm64)"
  echo ""
  echo "Internal (used by Dockerfile):"
  echo "  binary        Build binary for current platform (auto-detected)"
  echo ""
}

mkdir -p dist

case "$1" in
  docker)
    # Build for current platform only (local image for testing)
    BR=$(git branch --show-current)
    if [[ "${BR}" == "main" ]]; then
      BR=$(cat version)
    fi
    docker buildx build --load \
      -t ${TAG}:${BR} \
      -t ${TAG}:latest . || die "docker build failed"
    echo "Built ${TAG}:${BR} for $(uname -m)"
    ;;

  push)
    # Requires: docker buildx create --use
    BR=$(git branch --show-current)
    if [[ "${BR}" == "main" ]]; then
      BR=$(cat version)
    fi
    echo "Building multi-arch image for linux/amd64 and linux/arm64..."
    docker buildx build --platform linux/amd64,linux/arm64 \
      --provenance=false --sbom=false \
      -t ${TAG}:${BR} \
      -t ${TAG}:latest \
      --push . || die "docker build failed"
    echo "Pushed ${TAG}:${BR} (amd64 + arm64)"
    ;;

  binary)
    # Internal: called by Dockerfile. Builds for current arch (set by docker buildx)
    LDFLAGS="${LDFLAGS} -X main.docker=docker"
    env CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o keyhole main/keyhole.go
    echo "Built keyhole for $(go env GOOS)/$(go env GOARCH)"
    ;;

  binaries)
    echo "Building binaries for all platforms..."
    
    # Linux amd64
    env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/keyhole-linux-amd64 main/keyhole.go
    echo "  Built dist/keyhole-linux-amd64"
    
    # Linux arm64
    env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/keyhole-linux-arm64 main/keyhole.go
    echo "  Built dist/keyhole-linux-arm64"
    
    # macOS amd64
    env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/keyhole-darwin-amd64 main/keyhole.go
    echo "  Built dist/keyhole-darwin-amd64"
    
    # macOS arm64 (Apple Silicon)
    env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/keyhole-darwin-arm64 main/keyhole.go
    echo "  Built dist/keyhole-darwin-arm64"
    
    # Windows amd64
    env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/keyhole-windows-amd64.exe main/keyhole.go
    echo "  Built dist/keyhole-windows-amd64.exe"
    
    echo "Done! Binaries in dist/"
    ;;

  help|-h|--help)
    print_usage
    ;;

  "")
    rm -f dist/keyhole
    go build -ldflags "$LDFLAGS" -o dist/keyhole main/keyhole.go
    if [[ -f dist/keyhole ]]; then
      ./dist/keyhole -version
    fi
    ;;

  *)
    echo "Unknown command: $1"
    print_usage
    exit 1
    ;;
esac
