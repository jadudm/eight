variable "cf_password" {
  sensitive = true
}
variable "cf_username" {
  sensitive = true
}

variable cf_api_url {
  default = "https://api.fr.cloud.gov"
}

variable cf_env {
  default = "sandbox-gsa"
}

variable cf_org_name {
  default = "matthew.jadud"
}

variable api_key {
  sensitive = true
}


variable zap_debug_level {
  default = "info"
}

variable gin_debug_level {
  default = "release"
}