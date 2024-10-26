crawl:
	export VCAP_SERVICES=$(cat vcap_local.json) ; cd cmd/crawler ; go run *.go