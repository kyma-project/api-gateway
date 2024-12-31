variable "BTP_GLOBAL_ACCOUNT" {
  type        = string
  description = "Global account name"
}

variable "BTP_BOT_USER" {
  type        = string
  description = "Bot account name"
}

variable "BTP_BOT_PASSWORD" {
  type        = string
  description = "Bot account password"
}

variable "BTP_BACKEND_URL" {
  type        = string
  description = "BTP backend URL"
}

variable "BTP_NEW_SUBACCOUNT_NAME" {
  type        = string
  description = "Subaccount name"
}

variable "BTP_KYMA_PLAN" {
  type        = string
  description = "Plan name"
}

variable "BTP_NEW_SUBACCOUNT_REGION" {
  type        = string
  description = "Region name"
}

variable "BTP_CUSTOM_IAS_TENANT" {
  type        = string
  description = "Custom IAS tenant"
}

variable "BTP_KYMA_REGION" {
  type        = string
  description = "Kyma region"
}

variable "BTP_KYMA_MODULES_STRINGIFIED" {
  type        = string
  description = "Kyma modules as stringified json"
}

variable "BTP_KYMA_AUTOSCALER_MIN" {
  type = number
  default = 3
}

variable "BTP_KYMA_CUSTOM_ADMINISTRATORS" {
  type        = string
  description = "Emails of kyma admins stringified array"
}