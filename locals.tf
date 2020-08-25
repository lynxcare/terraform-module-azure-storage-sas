locals {
  expiration = timeadd(time_rotating.end.rotation_rfc3339, var.rotation_margin)
}