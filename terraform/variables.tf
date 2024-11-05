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

variable service_extract_ram {
  default = 320
}

variable service_fetch_ram {
  default = 128
}

variable service_pack_ram {
  default = 192
}

variable service_serve_ram {
  default = 128
}

variable service_walk_ram {
  default = 192
}