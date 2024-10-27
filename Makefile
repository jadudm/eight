build:
	docker build -t eight/dev -f Dockerfile.base .

run:
	docker compose up