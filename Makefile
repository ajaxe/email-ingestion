IMAGE_TAG ?= apogee-dev/email-ingestion:local
DOCKERFILE ?= infrastructure/Dockerfile
BUILD_CONTEXT ?= backend

.PHONY: docker-build

docker-build:
	docker build -f $(DOCKERFILE) -t $(IMAGE_TAG) $(BUILD_CONTEXT)
