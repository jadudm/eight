cleanup:
	cd cmd/pack ; rm -f *.sqlite ; rm -f *.sqlite-*

generate:
	cd internal/sqlite ; sqlc generate || exit 1

build: generate
	docker build -t eight/dev -f Dockerfile.base .

run: generate cleanup
	cd assets ; unzip -o static.zip ; echo 'static/*' > .gitignore
	docker compose up

cloc:
	docker run --rm -v ${PWD}:/tmp aldanial/cloc --exclude-dir=serve/static --exclude-dir=assets .