# Use official Python Alpine image
FROM python:3-alpine

# Install dos2unix and bash (bash needed if your script uses it)
RUN apk add --no-cache dos2unix bash

# Set working directory
WORKDIR /app

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh

# Convert line endings to Unix
RUN dos2unix /entrypoint.sh

# Make entrypoint executable
RUN chmod +x /entrypoint.sh

# Set container entrypoint
ENTRYPOINT ["/entrypoint.sh"]
