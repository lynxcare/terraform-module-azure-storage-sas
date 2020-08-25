output "sas" {
  description = "SAS token"
  value       = null_resource.rotation.outputs.sas
  sensitive   = true
}