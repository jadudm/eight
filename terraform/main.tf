module "database" {
  source = "github.com/gsa-tts/terraform-cloudgov//database?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "experiment-eight-db"
  recursive_delete = false
  tags             = ["eight"]
  rds_plan_name    = "micro-psql"
}

module "s3-private" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "experiment-eight-s3"
  s3_plan_name     = "basic"
  recursive_delete = false
  tags             = ["s3"]
}

data "cloudfoundry_domain" "public" {
  name = "app.cloud.gov"
}

data "cloudfoundry_space" "app_space" {
  org_name = "sandbox-gsa"
  name     = "matthew.jadud"
}

resource "cloudfoundry_route" "serve_route" {
  space    = data.cloudfoundry_space.app_space.id
  domain   = data.cloudfoundry_domain.public.id
  hostname = "experiment-eight"
}

# prepare one for each app... says how to deploy each app
# data "external" "fetch_zip" {
#   program     = ["python3", "scripts/prepare_fetch.py"]
#   working_dir = path.module
#   query = {
#     gitref = "refs/heads/main" # refs/where/branch
#   }
# }

resource "cloudfoundry_app" "fetch" {
  name                 = "fetch"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/fetch.zip"
  source_code_hash     = filesha256("zips/fetch.zip")
  disk_quota           = 128
  memory               = 64
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180

  service_binding {
    service_instance = module.s3-private.bucket_id
  }

  service_binding {
    service_instance = module.database.instance_id
  }

  # routes {
  #   route = cloudfoundry_route.app_route.id
  # }

  # Use for the first deployment
  environment = {
    ENV = "SANDBOX"
    # DISABLE_COLLECTSTATIC = 0
    DJANGO_BASE_URL    = "https://experiment-eight.app.cloud.gov"
    ALLOWED_HOSTS      = "experiment-eight.app.cloud.gov"
    REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}