terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Variables
variable "aws_region" {
  description = "AWS region"
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  default     = "production"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT signing secret"
  type        = string
  sensitive   = true
}

# VPC and Networking (simplified for MVP)
resource "aws_default_vpc" "default" {}

resource "aws_default_subnet" "default_az1" {
  availability_zone = "${var.aws_region}a"
}

resource "aws_default_subnet" "default_az2" {
  availability_zone = "${var.aws_region}b"
}

# Security Group for RDS
resource "aws_security_group" "rds" {
  name        = "jobscanner-rds-sg"
  description = "Security group for RDS PostgreSQL"
  vpc_id      = aws_default_vpc.default.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Restrict this in production
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# RDS PostgreSQL
resource "aws_db_subnet_group" "main" {
  name       = "jobscanner-db-subnet"
  subnet_ids = [aws_default_subnet.default_az1.id, aws_default_subnet.default_az2.id]
}

resource "aws_db_instance" "postgres" {
  identifier           = "jobscanner-db"
  engine              = "postgres"
  engine_version      = "16.1"
  instance_class      = "db.t3.micro"
  allocated_storage   = 20
  storage_type        = "gp2"
  
  db_name  = "jobscanner"
  username = "jobscanner"
  password = var.db_password
  
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  
  skip_final_snapshot = true
  publicly_accessible = true
  
  tags = {
    Name        = "jobscanner-db"
    Environment = var.environment
  }
}

# Outputs
output "db_endpoint" {
  value = aws_db_instance.postgres.endpoint
}

output "db_connection_string" {
  value     = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner"
  sensitive = true
}
