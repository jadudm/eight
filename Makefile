.PHONY: clean
clean:
	rm -rf assets/static
	rm -rf internal/sqlite/schemas/*.go
	rm -f cmd/*/service.exe

.PHONY: generate
generate:
	cd internal/sqlite ; sqlc generate || exit 1

docker: 
	docker build -t eight/dev -f Dockerfile.base .
	
.PHONY: build
build: clean generate
	cd cmd/extract ; make build
	cd cmd/fetch ; make build
	cd cmd/pack ; make build
	cd cmd/serve ; make build
	cd cmd/walk ; make build

.PHONY: run
run: generate clean
	cd assets ; unzip -o static.zip > /dev/null 2>&1
	docker compose up

.PHONY: terraform
terraform: build


.PHONY: cloc
cloc:
	docker run --rm -v ${PWD}:/tmp aldanial/cloc --exclude-dir=assets .