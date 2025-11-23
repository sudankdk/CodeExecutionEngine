FROM python:3

WORKDIR /app

# Copy your entrypoint script into the container
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set container entrypoint
ENTRYPOINT ["/entrypoint.sh"]
