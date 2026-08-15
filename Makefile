
.PHONY: docker-build docker-spa

docker-build:
	docker build --progress=plain -f infrastructure/Dockerfile -t apogee-dev/email-ingestion:local backend

docker-spa:
	docker build --progress=plain -f infrastructure/Dockerfile.spa -t apogee-dev/email-ingestion-spa:local frontend