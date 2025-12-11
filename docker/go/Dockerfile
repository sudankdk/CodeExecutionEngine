FROM golang:1.25-alpine

RUN addgroup -S runner && adduser -S runner -G runner

WORKDIR /app

RUN apk add --no-cache \
    bash \
    dos2unix

COPY entrypoint.sh /entrypoint.sh
RUN dos2unix /entrypoint.sh && chmod +x /entrypoint.sh

RUN chown -R runner:runner /app /entrypoint.sh

USER runner

ENV HOME=/home/runner
ENV GOPATH=/go

ENTRYPOINT ["/entrypoint.sh"]
