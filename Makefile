SHELL := C:/Program Files/Git/bin/bash.exe

.PHONY: docker-build docker-spa smtp-edge smtp-edge-upload smtp-edge-deploy apply-migration

docker-build:
	docker build --progress=plain -f infrastructure/Dockerfile -t apogee-dev/email-ingestion:local backend

docker-spa:
	docker build --progress=plain -f infrastructure/Dockerfile.spa -t apogee-dev/email-ingestion-spa:local frontend

smtp-edge: export CGO_ENABLED=0
smtp-edge: export GOOS=linux
smtp-edge: export GOARCH=arm64
smtp-edge:
	cd backend && go build -o ../infrastructure/smtp-edge-mta-service/output/email-ingest-arm64 main.go

smtp-edge-upload: smtp-edge
	pwsh -ExecutionPolicy Bypass -File infrastructure/smtp-edge-mta-service/upload.ps1

smtp-edge-deploy: smtp-edge-upload

apply-migration:
	cd backend && export $$(grep -v '^#' .env.dev | xargs 2>/dev/null) && go mod tidy && go build -o email-ingestion . && ./email-ingestion migrate up