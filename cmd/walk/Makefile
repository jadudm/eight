run:	
	go run *.go

test:
	go test

.PHONY: check
check:
	staticcheck

.PHONY: generate
generate:
	cd ../../internal/sqlite ; sqlc generate || exit 1

.PHONY: build
build: generate
	go build -buildvcs=false -o service.exe

clean:
	rm -f search.exe

terraform: clean build
	mkdir -p app
	echo "web: ./service.exe" > Procfile
	zip -r -X app.zip service.exe Procfile apt.yml
