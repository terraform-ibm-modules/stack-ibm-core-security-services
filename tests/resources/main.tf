##############################################################################
# Resource Group
##############################################################################

module "resource_group" {
  source  = "terraform-ibm-modules/resource-group/ibm"
  version = "1.1.6"
  # if an existing resource group is not set (null) create a new one using prefix
  resource_group_name          = var.resource_group == null ? "${var.prefix}-resource-group" : null
  existing_resource_group_name = var.resource_group
}

##############################################################################
# Event Notification
##############################################################################

module "event_notifications" {
  source            = "terraform-ibm-modules/event-notifications/ibm"
  version           = "1.10.5"
  resource_group_id = module.resource_group.resource_group_id
  name              = "${var.prefix}-en"
  tags              = var.resource_tags
  plan              = "lite"
  service_endpoints = "public"
  region            = var.region
}

##############################################################################
# Secrets Manager
##############################################################################

module "secrets_manager" {
  source               = "terraform-ibm-modules/secrets-manager/ibm"
  version              = "1.17.9"
  resource_group_id    = module.resource_group.resource_group_id
  region               = var.region
  secrets_manager_name = "${var.prefix}-secrets-manager" #tfsec:ignore:general-secrets-no-plaintext-exposure
  sm_service_plan      = "trial"
  sm_tags              = var.resource_tags
}

#############################################################################
# Provision cloud object storage and bucket
#############################################################################

module "cos" {
  source                 = "terraform-ibm-modules/cos/ibm"
  version                = "8.10.1"
  resource_group_id      = module.resource_group.resource_group_id
  region                 = var.region
  cross_region_location  = null
  cos_instance_name      = "${var.prefix}-vpc-logs-cos"
  cos_tags               = var.resource_tags
  bucket_name            = "${var.prefix}-vpc-logs-cos-bucket"
  kms_encryption_enabled = false
  retention_enabled      = false
}

##############################################################################
# SCC
##############################################################################

module "create_scc_instance" {
  source                            = "terraform-ibm-modules/scc/ibm"
  version                           = "1.7.2"
  instance_name                     = "${var.prefix}-scc-instance"
  region                            = var.region
  resource_group_id                 = module.resource_group.resource_group_id
  resource_tags                     = var.resource_tags
  access_tags                       = []
  cos_bucket                        = module.cos.bucket_name
  cos_instance_crn                  = module.cos.cos_instance_id
  attach_wp_to_scc_instance         = false
  skip_cos_iam_authorization_policy = false
}
