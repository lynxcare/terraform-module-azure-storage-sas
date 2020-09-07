terraform {
  experiments      = [variable_validation] // added for Terraform 0.12
  required_version = ">= 0.12.29"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 2.24.0"
    }
    time = {
      source  = "hashicorp/time"
      version = ">= 0.5.0"
    }
  }
}