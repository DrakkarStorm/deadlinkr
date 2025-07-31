#!/bin/bash
# ==============================================================================
# Docker Build Script for Deadlinkr
# ==============================================================================

set -euo pipefail

# Configuration
IMAGE_NAME="deadlinkr"
REGISTRY="ghcr.io"
REPO_NAME="drakkarstorm/deadlinkr"
BUILD_CONTEXT="."
DOCKERFILE="Dockerfile"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
🐳 Docker Build Script for Deadlinkr

Usage: $0 [OPTIONS]

OPTIONS:
    -t, --tag TAG           Tag for the Docker image (default: latest)
    -r, --registry REGISTRY Registry to push to (default: ghcr.io)
    -p, --push              Push image to registry after build
    -m, --multi-arch        Build for multiple architectures (linux/amd64,linux/arm64)
    -c, --clean             Clean build (no cache)
    -s, --scan              Run security scan after build
    -h, --help              Show this help message

EXAMPLES:
    # Build local image
    $0 --tag v1.0.0

    # Build and push to registry
    $0 --tag v1.0.0 --push

    # Build multi-architecture image
    $0 --tag v1.0.0 --multi-arch --push


    # Clean build with security scan
    $0 --clean --scan --tag latest

EOF
}

# Parse command line arguments
TAG="latest"
REGISTRY_URL="$REGISTRY"
PUSH=false
MULTI_ARCH=false
CLEAN=false
SCAN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY_URL="$2"
            shift 2
            ;;
        -p|--push)
            PUSH=true
            shift
            ;;
        -m|--multi-arch)
            MULTI_ARCH=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -s|--scan)
            SCAN=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate Docker is running
if ! docker info >/dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Set up buildx if multi-arch is requested
if [[ "$MULTI_ARCH" == true ]]; then
    log_info "Setting up Docker Buildx for multi-architecture build..."
    docker buildx create --name deadlinkr-builder --use 2>/dev/null || true
    docker buildx inspect --bootstrap
fi

# Determine image tags
LOCAL_TAG="${IMAGE_NAME}:${TAG}"
REMOTE_TAG="${REGISTRY_URL}/${REPO_NAME}:${TAG}"

# Build arguments
BUILD_ARGS=()
BUILD_ARGS+=("--file" "$DOCKERFILE")
BUILD_ARGS+=("--tag" "$LOCAL_TAG")

if [[ "$PUSH" == true ]]; then
    BUILD_ARGS+=("--tag" "$REMOTE_TAG")
fi

if [[ "$CLEAN" == true ]]; then
    BUILD_ARGS+=("--no-cache")
fi

if [[ "$MULTI_ARCH" == true ]]; then
    BUILD_ARGS+=("--platform" "linux/amd64,linux/arm64")
    if [[ "$PUSH" == true ]]; then
        BUILD_ARGS+=("--push")
    fi
else
    if [[ "$PUSH" == true ]]; then
        BUILD_ARGS+=("--push")
    else
        BUILD_ARGS+=("--load")
    fi
fi

# Add build context
BUILD_ARGS+=("$BUILD_CONTEXT")

# Display build configuration
log_info "🐳 Docker Build Configuration"
echo "  Image: $LOCAL_TAG"
if [[ "$PUSH" == true ]]; then
    echo "  Remote: $REMOTE_TAG"
fi
echo "  Dockerfile: $DOCKERFILE"
echo "  Multi-arch: $MULTI_ARCH"
echo "  Clean build: $CLEAN"
echo "  Push: $PUSH"
echo "  Security scan: $SCAN"
echo ""

# Start build
log_info "🏗️ Starting Docker build..."

if [[ "$MULTI_ARCH" == true ]]; then
    docker buildx build "${BUILD_ARGS[@]}"
else
    docker build "${BUILD_ARGS[@]}"
fi

if [[ $? -eq 0 ]]; then
    log_success "✅ Docker image built successfully: $LOCAL_TAG"
else
    log_error "❌ Docker build failed"
    exit 1
fi

# Display image information
if [[ "$MULTI_ARCH" == false && "$PUSH" == false ]]; then
    log_info "📊 Image information:"
    docker images "$LOCAL_TAG" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
fi

# Security scan
if [[ "$SCAN" == true ]]; then
    log_info "🛡️ Running security scan..."
    
    if command -v trivy >/dev/null 2>&1; then
        trivy image "$LOCAL_TAG"
    else
        log_warning "Trivy not found. Installing via Docker..."
        docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
            aquasec/trivy:latest image "$LOCAL_TAG"
    fi
fi

# Test the image
if [[ "$MULTI_ARCH" == false && "$PUSH" == false ]]; then
    log_info "🧪 Testing the Docker image..."
    
    if docker run --rm "$LOCAL_TAG" --help >/dev/null 2>&1; then
        log_success "✅ Image test passed"
    else
        log_error "❌ Image test failed"
        exit 1
    fi
fi

log_success "🎉 Build completed successfully!"

# Show usage examples
echo ""
log_info "📖 Usage examples:"
echo "  # Run a scan"
echo "  docker run --rm $LOCAL_TAG scan https://example.com"
echo ""
echo "  # Run with authentication"
echo "  docker run --rm $LOCAL_TAG scan https://private-site.com --auth-basic 'user:pass'"
echo ""
echo "  # Run with custom configuration"
echo "  docker run --rm -v \$(pwd)/config:/app/config $LOCAL_TAG scan https://example.com --depth 3"