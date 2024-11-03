run:	
	go run *.go

test:
	go test

check:
	staticcheck

generate:
	cd ../../internal/sqlite ; sqlc generate || exit 1

build: generate
	go build -buildvcs=false -o service.exe

clean:
	rm -f fetch.exe
	rm -f Procfile
	rm -f apt.yml
