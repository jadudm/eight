run:	
	go run *.go

test:
	go test

check:
	staticcheck

build: generate
	go build --tags fts5 -buildvcs=false -o service.exe

clean:
	rm -f fetch.exe
	rm -f Procfile
	rm -f apt.yml

terraform: clean build
	mkdir -p app
	echo "web: ./service.exe" > Procfile
	zip -r -X app.zip service.exe Procfile apt.yml

