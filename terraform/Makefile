.PHONY: init
init:
	terraform init

.PHONY: clean
clean:
	rm -rf .terraform
	rm -f .terraform.lock.hcl
	rm -rf app/
	rm -rf zips/*.zip
	mkdir -p plans

.PHONY: clean_app
clean_app:
	rm -rf app/

.PHONY: app
app: clean_app
	mkdir -p app
	cd .. ; \
		cp -r assets terraform/app/ ; \
		cp -r config terraform/app/

.PHONY: extract
extract: app
	rm -f app/*.exe
	cd ../cmd/extract ; make build
	cp ../cmd/extract/service.exe app/extract.exe
	echo "web: ./extract.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/extract.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x assetts/static/* \
		-x terraform/*

.PHONY: fetch
fetch: app
	rm -f app/*.exe
	cd ../cmd/fetch ; make build
	cp ../cmd/fetch/service.exe app/fetch.exe
	echo "web: ./fetch.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/fetch.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x assetts/static/* \
		-x terraform/*

.PHONY: pack
pack: app
	rm -f app/*.exe
	cd ../cmd/pack ; make build
	cp ../cmd/pack/service.exe app/pack.exe
	echo "web: ./pack.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/pack.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x assets/static/* \
		-x terraform/*

.PHONY: serve
serve: app
	rm -f app/*.exe
	cd ../cmd/serve ; make build
	cp ../cmd/serve/service.exe app/serve.exe
	echo "web: ./serve.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cp ../assets/static.zip app/static.zip
	cd app ; unzip static.zip ; rm static.zip 
	cp ../cmd/serve/index.html app/static/index.html
	mkdir app/databases ;
	cd app ; zip -r -X ../zips/serve.zip . \
		-x assets/databases/*sqlite* \
		-x assets/static.zip \
		-x terraform/*

.PHONY: walk
walk: app
	rm -f app/*.exe
	cd ../cmd/walk ; make build
	cp ../cmd/walk/service.exe app/walk.exe
	echo "web: ./walk.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/walk.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x assetts/static/* \
		-x terraform/*

.PHONY:
plan:
	terraform plan -var-file="terraform.tfvars" -out plans/das.plan

.PHONY: tf_extract
tf_extract: extract plan apply
	echo "tf extract"

.PHONY: tf_fetch
tf_fetch: plan apply
	echo "tf fetch"

.PHONY: tf_pack
tf_pack: pack plan apply
	echo "tf pack"

.PHONY: tf_serve
tf_serve: serve plan apply
	echo "tf serve"

.PHONY: tf_walk
tf_walk: walk plan apply
	echo "tf walk"
	
.PHONY: plan_all
plan_all: clean init extract fetch pack serve walk
	terraform plan -var-file="terraform.tfvars" -out plans/das.plan

.PHONY: apply
apply:
	terraform apply plans/das.plan

.PHONY: apply_all
apply_all: clean plan_all apply
	echo "apply_all"

