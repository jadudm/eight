cleanup:
	cd cmd/pack ; rm -f *.sqlite ; rm -f *.sqlite-journal

generate:
	cd internal/sqlite ; sqlc generate || exit 1

build: generate
	docker build -t eight/dev -f Dockerfile.base .

run: generate cleanup
	docker compose up