# Use official Node.js LTS Alpine image
FROM node:20-alpine

# Install dos2unix and bash
RUN apk add --no-cache dos2unix bash

# Set working directory inside container
WORKDIR /app

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh

# Convert line endings to Unix
RUN dos2unix /entrypoint.sh

# Make entrypoint executable
RUN chmod +x /entrypoint.sh

# Use the entrypoint script
ENTRYPOINT ["/entrypoint.sh"]
