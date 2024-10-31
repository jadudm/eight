run:	
	go run *.go

generate:
	cd ../../internal/sqlite ; sqlc generate || exit 1

build: generate
	go build -buildvcs=false -o fetch.exe

test:
	go test

check:
	staticcheck
