PORT ?= 8888

# make run
#
# or
#
# make run PORT=3000
#
# to change the port of the server. 

run:
	DATABASE_URL=postgresql://postgres@localhost:5432/postgres \
	go run *.go -p ${PORT}