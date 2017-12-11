.PHONY: all build release

IMAGE=dddpaul/vscalebot
VERSION=0.7

all: build

build-alpine:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/vscalebot ./main.go

build:
	@docker build --tag=${IMAGE} .

debug:
	@docker run -it --entrypoint=sh ${IMAGE}

release: build
	@docker build --tag=${IMAGE}:${VERSION} .

deploy: release
	@docker push ${IMAGE}
	@docker push ${IMAGE}:${VERSION}
