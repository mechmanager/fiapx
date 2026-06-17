variable "aws_region" {
  description = "Região AWS onde o cluster EKS será criado"
  type        = string
  default     = "us-east-1"
}

variable "cluster_name" {
  description = "Nome do cluster EKS"
  type        = string
  default     = "fiapx"
}

variable "cluster_version" {
  description = "Versão do Kubernetes no EKS"
  type        = string
  default     = "1.31"
}

variable "vpc_cidr" {
  description = "CIDR block da VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "node_instance_type" {
  description = "Tipo de instância EC2 dos nós workers"
  type        = string
  default     = "t3.medium"
}

variable "node_min_size" {
  description = "Número mínimo de nós no node group"
  type        = number
  default     = 2
}

variable "node_max_size" {
  description = "Número máximo de nós no node group"
  type        = number
  default     = 10
}

variable "node_desired_size" {
  description = "Número desejado de nós no node group"
  type        = number
  default     = 3
}

variable "tags" {
  description = "Tags aplicadas a todos os recursos"
  type        = map(string)
  default = {
    Project     = "fiapx"
    ManagedBy   = "terraform"
    Environment = "production"
  }
}
