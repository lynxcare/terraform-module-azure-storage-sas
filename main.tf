provider "azurerm" {
  features {}
}

resource "time_static" "start" {
  rfc3339 = coalesce(var.start, timestamp())
}

resource "time_rotating" "end" {
  rfc3339          = time_static.start.rfc3339
  rotation_days    = var.rotation_days
  rotation_hours   = var.rotation_hours
  rotation_minutes = var.rotation_minutes
  rotation_months  = var.rotation_months
  rotation_years   = var.rotation_years
}

data "azurerm_storage_account" "sa" {
  name                = var.storage_account_name
  resource_group_name = var.resource_group_name
}

data "azurerm_storage_account_sas" "sas" {
  count             = var.storage_container_name == null || var.storage_container_name == "" ? 1 : 0
  connection_string = data.azurerm_storage_account.sa.primary_connection_string
  expiry            = local.expiration
  start             = time_static.start.rfc3339
  permissions {
    add     = var.write
    create  = var.write
    delete  = var.write
    list    = true
    process = var.write
    read    = true
    update  = var.write
    write   = var.write
  }
  resource_types {
    container = true
    object    = true
    service   = true
  }
  services {
    blob  = true
    file  = true
    queue = true
    table = true
  }
}

data "azurerm_storage_account_blob_container_sas" "sas" {
  count             = var.storage_container_name == null || var.storage_container_name == "" ? 0 : 1
  connection_string = data.azurerm_storage_account.sa.primary_blob_connection_string
  container_name    = var.storage_container_name
  expiry            = local.expiration
  start             = time_static.start.rfc3339
  permissions {
    add    = var.write
    create = var.write
    delete = var.write
    list   = true
    read   = true
    write  = var.write
  }
}

resource "null_resource" "rotation" {
  triggers = {
    start_time = time_static.start.rfc3339
    end_time   = local.expiration
  }

  inputs = {
    sas = var.storage_container_name == null || var.storage_container_name == "" ? data.azurerm_storage_account_sas.sas[0].sas : data.azurerm_storage_account_blob_container_sas.sas[0].sas
  }
}