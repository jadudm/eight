.PHONY: init
init:
	terraform init

.PHONY: clean
clean:
	rm -rf .terraform
	rm -f .terraform.lock.hcl
	rm -rf app/
	mkdir -p plans

.PHONY: app
app:
	mkdir -p app
	cd .. ; \
		cp -r assets terraform/app/ ; \
		cp -r config terraform/app/

.PHONY: clean_fetch
clean_fetch: 
	rm -f zips/fetch.zip

.PHONY: fetch
fetch: clean_fetch app
	cp ../cmd/fetch/service.exe app/fetch.exe
	echo "web: ./app/fetch.exe" > app/Procfile
	cp ../cmd/fetch/apt.yml app/apt.yml
	cd app ; zip -r -X ../zips/fetch.zip . \
		-x assets/databases/* \
		-x assets/static.zip

.PHONY: plan
plan: clean init fetch
	terraform plan -var-file="terraform.tfvars" -out plans/das.plan

.PHONY: apply
apply: plan
	terraform apply plans/das.plan