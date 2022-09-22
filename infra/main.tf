terraform {
  backend "azurerm" {
    resource_group_name  = "tfstate"
    storage_account_name = "tmtfstate"
    container_name       = "tfstate"
    key                  = "foodlog.terraform.tfstate"
  }

  required_providers {
    azapi = {
      source = "Azure/azapi"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=3.7.0"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azapi" {
}

variable "foodlog_web_image" {
  type = string
}

resource "azurerm_resource_group" "rg" {
  name     = "foodlog"
  location = "uksouth"
}

resource "azurerm_storage_account" "stg" {
  name                = "foodlogstore0"
  location            = "uksouth"
  resource_group_name = azurerm_resource_group.rg.name

  account_kind             = "StorageV2"
  account_replication_type = "LRS"
  account_tier             = "Standard"
}

resource "azurerm_storage_container" "db_backups" {
  name                 = "db-backups"
  storage_account_name = azurerm_storage_account.stg.name
}

data "azurerm_storage_account_blob_container_sas" "backup_sas" {
  connection_string = azurerm_storage_account.stg.primary_connection_string
  container_name    = azurerm_storage_container.db_backups.name
  https_only        = true

  start  = "2022-07-31"
  expiry = "2023-07-31"

  permissions {
    read   = true
    add    = true
    create = true
    write  = true
    delete = true
    list   = true
  }

  cache_control       = "max-age=5"
  content_disposition = "inline"
  content_encoding    = "deflate"
  content_language    = "en-US"
  content_type        = "application/json"
}

resource "azurerm_container_registry" "acr" {
  name                = "foodlogacr0"
  location            = "uksouth"
  resource_group_name = azurerm_resource_group.rg.name

  sku           = "Standard"
  admin_enabled = true
}

resource "azurerm_log_analytics_workspace" "logs" {
  name                = "foodloglogs"
  location            = "uksouth"
  resource_group_name = azurerm_resource_group.rg.name

  retention_in_days = 30
}

resource "azapi_resource" "managed_environment" {
  name      = "foodlogenv"
  location  = "uksouth"
  parent_id = azurerm_resource_group.rg.id
  type      = "Microsoft.App/managedEnvironments@2022-03-01"

  body = jsonencode({
    properties = {
      appLogsConfiguration = {
        destination = "log-analytics"
        logAnalyticsConfiguration = {
          customerId = azurerm_log_analytics_workspace.logs.workspace_id
          sharedKey  = azurerm_log_analytics_workspace.logs.primary_shared_key
        }
      }
    }
  })
}

resource "azapi_resource" "container_app" {
  name      = "foodlogweb"
  location  = "uksouth"
  parent_id = azurerm_resource_group.rg.id
  type      = "Microsoft.App/containerApps@2022-03-01"

  body = jsonencode({
    properties : {
      managedEnvironmentId = azapi_resource.managed_environment.id
      configuration = {
        ingress = {
          external   = true
          targetPort = 8080
        }
        registries = [{
          server            = azurerm_container_registry.acr.login_server
          username          = azurerm_container_registry.acr.admin_username
          passwordSecretRef = "docker-password"
        }]
        secrets = [
          {
            name  = "docker-password"
            value = azurerm_container_registry.acr.admin_password
          },
          {
            name  = "backup-sas"
            value = "${azurerm_storage_account.stg.primary_blob_endpoint}${azurerm_storage_container.db_backups.name}/${data.azurerm_storage_account_blob_container_sas.backup_sas.sas}"
          }
        ]
      }
      template = {
        containers = [{
          image = "${var.foodlog_web_image}"
          name  = "foodlogweb"
          env = [{
            name      = "FOODLOG_BACKUP_SAS"
            secretRef = "backup-sas"
          }]
          resources = {
            cpu : 0.25
            memory : "0.5Gi"
          }
        }]
        scale = {
          minReplicas = 1
          maxReplicas = 1
        }
      }
    }
  })
}

output "backup_sas_url" {
  value     = "${azurerm_storage_account.stg.primary_blob_endpoint}${azurerm_storage_container.db_backups.name}/${data.azurerm_storage_account_blob_container_sas.backup_sas.sas}"
  sensitive = true
}