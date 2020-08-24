terraform {
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