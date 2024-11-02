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
	cp ../cmd/extract/service.exe app/extract.exe
	echo "web: ./extract.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/extract.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x terraform/*

.PHONY: fetch
fetch: app
	rm -f app/*.exe
	cp ../cmd/fetch/service.exe app/fetch.exe
	echo "web: ./fetch.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/fetch.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x terraform/*

.PHONY: pack
pack: app
	rm -f app/*.exe
	cp ../cmd/pack/service.exe app/pack.exe
	echo "web: ./pack.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/pack.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x terraform/*

.PHONY: serve
serve: app
	rm -f app/*.exe
	cp ../cmd/serve/service.exe app/serve.exe
	echo "web: ./serve.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/serve.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x terraform/*

.PHONY: walk
walk: app
	rm -f app/*.exe
	cp ../cmd/walk/service.exe app/walk.exe
	echo "web: ./walk.exe" > app/Procfile
	cp ../assets/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/walk.zip . \
		-x assets/databases/* \
		-x assets/static.zip \
		-x terraform/*

.PHONY: plan
plan: clean init extract fetch pack serve walk
	terraform plan -var-file="terraform.tfvars" -out plans/das.plan

.PHONY: apply
apply: clean plan
	terraform apply plans/das.plan