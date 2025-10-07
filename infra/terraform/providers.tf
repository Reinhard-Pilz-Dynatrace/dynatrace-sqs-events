terraform {
  required_version = ">= 1.3.0"
  required_providers {
    aws    = { source = "hashicorp/aws", version = ">= 5.0" }
    http   = { source = "hashicorp/http", version = ">= 3.4" }
    github = { source = "integrations/github", version = ">= 6.0" }
    local = { source = "hashicorp/local", version = ">= 2.4" }
  }
}

provider "github" {
  owner = var.github_owner
  token       = var.github_token  
}

provider "local" {}

provider "aws" {
  region = var.region
}