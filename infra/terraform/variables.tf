# Required variables
variable "region" { type = string }
variable "queue_arn" { type = string }

variable "function_name" { type = string }

variable "dt_url" { type = string }
variable "dt_token" { type = string }

variable "architecture" { default = "arm64" }

variable "dt_entity_selector" { default = "" }
variable "dt_event_type" { default = "AVAILABILITY_EVENT" }
variable "dt_timeout_ms" { default = 0 }
variable "dt_title_max" { default = 500 }
variable "dt_original_queue_prop" { default = "originalQueueArn" }
variable "dt_dlq_prefix" { default = "[DQL]" }

variable "github_owner" { default = "Reinhard-Pilz-Dynatrace" }
variable "github_repo"  { default = "dynatrace-sqs-events" }
variable "github_token" { default = "" }