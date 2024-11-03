run:	
	go run *.go

test:
	go test

check:
	staticcheck

generate:
	cd ../../internal/sqlite ; sqlc generate || exit 1

# -tags=viper_bind_struct 
build: generate
	go build -buildvcs=false -o service.exe

clean:
	rm -f service.exe
