locals {
  expiration = timeadd(time_rotating.end.rfc3339, var.rotation_margin)
}