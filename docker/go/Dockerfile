# Stage 1: Base image with Go preinstalled
FROM golang:1.25-alpine

# Install only what's needed for sandboxing
RUN apk add --no-cache \
    bash \
    libc-dev \
    busybox-extras \
    tini


WORKDIR /app

# Copy the sandbox runner script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Use tini as init to handle signals correctly
ENTRYPOINT ["/entrypoint.sh"]
