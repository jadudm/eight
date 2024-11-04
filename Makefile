.PHONY: clean
clean:
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
	cd assets ; rm -rf static ; unzip static.zip
	cd cmd/serve ; cp index.html ../../assets/
	cd cmd/serve ; cp index.html ../../assets/static

.PHONY: up
up: build
	docker compose -f compiled.yaml up

.PHONY: run
run: clean generate
	cd assets ; unzip -o static.zip > /dev/null 2>&1
	docker compose up

.PHONY: terraform
terraform: build
	cd terraform ; make apply_all

.PHONY: cloc
cloc:
	docker run --rm -v ${PWD}:/tmp aldanial/cloc --exclude-dir=assets .

delete_extract:
	cf delete -f extract

delete_fetch:
	cf delete -f fetch

delete_pack:
	cf delete -f pack

delete_serve:
	cf delete -f serve

delete_walk:
	cf delete -f walk

.PHONY: delete_all
delete_all: delete_extract delete_fetch delete_pack delete_serve delete_walk
