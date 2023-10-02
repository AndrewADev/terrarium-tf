provider "google" {
  region = "eu-central-1"
}

data "aws_caller_identity" "self" {}

variable "region" {}
variable "environment" {}
variable "project" {}
variable "account" {}
variable "stack" {}
variable "foo" {
  type = bool
}

resource "google_storage_bucket" "test" {
  bucket = "test-${var.environment}-${data.aws_caller_identity.self.account_id}"
}

resource "google_storage_bucket" "state" {}

terraform {
  backend "gcs"  {
  }
}

output "foo" {
  value = "test-${var.environment}-${data.aws_caller_identity.self.account_id}"
}
