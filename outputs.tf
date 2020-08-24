output "sas" {
  description = "SAS token"
  value       = var.storage_container_name == null || var.storage_container_name == "" ? data.azurerm_storage_account_sas.sas.sas : data.azurerm_storage_account_blob_container_sas.sas.sas
  sensitive   = true
}