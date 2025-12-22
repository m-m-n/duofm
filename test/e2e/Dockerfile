# syntax=docker/dockerfile:1.6

# E2E Test Environment for duofm
# Tests file manager behavior including permission handling
FROM golang:1.23-bookworm

# Install tmux and other utilities with security updates
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends \
        tmux \
        ncurses-bin \
        less \
        vim && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Allow downloading newer toolchain if needed
ENV GOTOOLCHAIN=auto

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the application with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /usr/local/bin/duofm ./cmd/duofm

# Create non-root test user
RUN useradd -m -u 1000 -s /bin/bash testuser

# Create test fixtures directory with mixed ownership
RUN mkdir -p /testdata && \
    mkdir -p /testdata/dir1/subdir && \
    mkdir -p /testdata/dir2 && \
    mkdir -p /testdata/empty_dir && \
    mkdir -p /testdata/user_owned && \
    mkdir -p /testdata/root_owned && \
    mkdir -p /testdata/no_access && \
    # Files owned by testuser (can be deleted)
    echo "user file content" > /testdata/user_owned/deletable.txt && \
    chown -R testuser:testuser /testdata/user_owned && \
    # Files owned by root (cannot be deleted by testuser)
    echo "root file content" > /testdata/root_owned/protected.txt && \
    chmod 644 /testdata/root_owned/protected.txt && \
    # Directory with no access permission
    chmod 000 /testdata/no_access && \
    # Regular test files (owned by root, readable)
    echo "file1 content" > /testdata/file1.txt && \
    echo "file2 content" > /testdata/file2.txt && \
    echo "subfile content" > /testdata/dir1/subdir/subfile.txt && \
    echo "another file" > /testdata/dir2/another.txt && \
    ln -s /testdata/file1.txt /testdata/link_to_file1.txt && \
    # Permission denied directory for navigation test
    chmod 000 /testdata/dir1/subdir

# Copy test scripts and make executable for all users
COPY test/e2e/scripts /e2e/scripts
RUN chmod -R 755 /e2e/scripts

WORKDIR /testdata

# Run as non-root user to test permission handling
USER testuser

# Default command runs E2E tests
CMD ["/e2e/scripts/run_tests.sh"]
