# SHELL := /bin/bash
# FIXME: Move to config
PORT ?= 8888

# make run
#
# or
#
# make run PORT=3000
#
# to change the port of the server. 
run:	
	# FIXME: Move to config
	#cp ../../vcap.json . 
	#DATABASE_URL=postgresql://postgres@localhost:5432/postgres \
	go run *.go -p ${PORT}

build: 
	go build -buildvcs=false -o crawler.exe

test:
	go test

run:
	go run *.go

check:
	staticcheck
