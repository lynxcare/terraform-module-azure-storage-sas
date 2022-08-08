output "sas" {
  description = "SAS token"
  value       = var.storage_container_name == null || var.storage_container_name == "" ? data.azurerm_storage_account_sas.sas[0].sas : data.azurerm_storage_account_blob_container_sas.sas[0].sas
  sensitive   = true
}

output "expiry_date" {
  description = "expiry date"
  value       = local.expiration
}