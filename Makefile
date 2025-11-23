build-python:
	docker build -t cee-python ./docker/python

build-node:
	docker build -t cee-node ./docker/node

build-go:
	docker build -t cee-go ./docker/go

build-all: build-python build-node build-go


checker-node: docker run --rm -it --entrypoint ls cee-node -l /entrypoint.sh