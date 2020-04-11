.PHONY: all node

IMAGE_NAME=$(if $(ENV_IMAGE_NAME),$(ENV_IMAGE_NAME),hydro-monitor/node)
IMAGE_VERSION=$(if $(ENV_IMAGE_VERSION),$(ENV_IMAGE_VERSION),v0.0.0)

$(info node image settings: $(IMAGE_NAME) version $(IMAGE_VERSION))

all: node

test:
	go test github.com/hydro-monitor/node/pkg/... -cover
	go vet github.com/hydro-monitor/node/pkg/...

node:
	go build -o _output/node ./cmd

image-node:
	go mod vendor
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) -f deploy/docker/Dockerfile .

push-image-node: image-node
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)

clean:
	go clean -r -x
	rm -f deploy/docker/node
