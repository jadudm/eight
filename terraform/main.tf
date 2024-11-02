
data "cloudfoundry_domain" "public" {
  name = "app.cloud.gov"
}

data "cloudfoundry_space" "app_space" {
  org_name = "sandbox-gsa"
  name     = "matthew.jadud"
}

#################################################################
# POSTGRES
#################################################################

module "database" {
  source = "github.com/gsa-tts/terraform-cloudgov//database?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "experiment-eight-db"
  recursive_delete = false
  tags             = ["eight"]
  rds_plan_name    = "micro-psql"
}

#################################################################
# S3 BUCKETS
#################################################################
module "s3-private-extract" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "extract"
  s3_plan_name     = "basic"
  recursive_delete = false
  tags             = ["s3"]
}

module "s3-private-fetch" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "fetch"
  s3_plan_name     = "basic"
  recursive_delete = false
  tags             = ["s3"]
}

module "s3-private-serve" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "serve"
  s3_plan_name     = "basic"
  recursive_delete = false
  tags             = ["s3"]
}


#################################################################
# FETCH
#################################################################

resource "cloudfoundry_route" "fetch_route" {
  space    = data.cloudfoundry_space.app_space.id
  domain   = data.cloudfoundry_domain.public.id
  hostname = "fetch-experiment-eight"
}
resource "cloudfoundry_app" "fetch" {
  name                 = "fetch"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/fetch.zip"
  source_code_hash     = filesha256("zips/fetch.zip")
  disk_quota           = 512
  memory               = 64
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
    service_binding {
    service_instance = module.s3-private-fetch.bucket_id
  }

  service_binding {
    service_instance = module.database.instance_id
  }

  routes {
    route = cloudfoundry_route.fetch_route.id
  }

  environment = {
    ENV = "SANDBOX"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}

#################################################################
# EXTRACT
#################################################################
resource "cloudfoundry_app" "extract" {
  name                 = "extract"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/extract.zip"
  source_code_hash     = filesha256("zips/extract.zip")
  disk_quota           = 512
  memory               = 128
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
  service_binding {
    service_instance = module.s3-private-extract.bucket_id
  }
  service_binding {
    service_instance = module.s3-private-fetch.bucket_id
  }
  service_binding {
    service_instance = module.database.instance_id
  }

  # routes {
  #   route = cloudfoundry_route.serve_route.id
  # }

  environment = {
    ENV = "SANDBOX"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}

#################################################################
# PACK
#################################################################

resource "cloudfoundry_app" "pack" {
  name                 = "pack"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/pack.zip"
  source_code_hash     = filesha256("zips/pack.zip")
  disk_quota           = 512
  memory               = 128
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
  service_binding {
    service_instance = module.s3-private-extract.bucket_id
  }
  service_binding {
    service_instance = module.s3-private-serve.bucket_id
  }

  service_binding {
    service_instance = module.database.instance_id
  }

  # routes {
  #   route = cloudfoundry_route.serve_route.id
  # }

  environment = {
    ENV = "SANDBOX"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}


#################################################################
# SERVE
#################################################################

resource "cloudfoundry_route" "serve_route" {
  space    = data.cloudfoundry_space.app_space.id
  domain   = data.cloudfoundry_domain.public.id
  hostname = "serve-experiment-eight"
}
resource "cloudfoundry_app" "serve" {
  name                 = "serve"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/serve.zip"
  source_code_hash     = filesha256("zips/serve.zip")
  disk_quota           = 512
  memory               = 64
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
    service_binding {
    service_instance = module.s3-private-fetch.bucket_id
  }
  service_binding {
    service_instance = module.s3-private-serve.bucket_id
  }
  service_binding {
    service_instance = module.database.instance_id
  }

  routes {
    route = cloudfoundry_route.serve_route.id
  }

  environment = {
    ENV = "SANDBOX"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}

#################################################################
# WALK
#################################################################

resource "cloudfoundry_app" "walk" {
  name                 = "walk"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/walk.zip"
  source_code_hash     = filesha256("zips/walk.zip")
  disk_quota           = 512
  memory               = 64
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
  service_binding {
    service_instance = module.s3-private-fetch.bucket_id
  }

  service_binding {
    service_instance = module.database.instance_id
  }

  # routes {
  #   route = cloudfoundry_route.serve_route.id
  # }

  environment = {
    ENV = "SANDBOX"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}