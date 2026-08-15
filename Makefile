
.PHONY: docker-build docker-spa smtp-edge

docker-build:
	docker build --progress=plain -f infrastructure/Dockerfile -t apogee-dev/email-ingestion:local backend

docker-spa:
	docker build --progress=plain -f infrastructure/Dockerfile.spa -t apogee-dev/email-ingestion-spa:local frontend

smtp-edge: export CGO_ENABLED=0
smtp-edge: export GOOS=linux
smtp-edge: export GOARCH=arm64
smtp-edge:
	cd backend && go build -o ../output/email-ingest-arm64 main.go