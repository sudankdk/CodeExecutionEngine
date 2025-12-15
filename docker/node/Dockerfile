FROM node:20-alpine

RUN addgroup -S runner && adduser -S runner -G runner

WORKDIR /app

RUN apk add --no-cache bash dos2unix

# Create the /run/code directory
RUN mkdir -p /run/code && chown -R runner:runner /run/code

COPY entrypoint.sh /entrypoint.sh
RUN dos2unix /entrypoint.sh && chmod +x /entrypoint.sh

RUN chown -R runner:runner /app /entrypoint.sh

USER runner

ENTRYPOINT ["/entrypoint.sh"]