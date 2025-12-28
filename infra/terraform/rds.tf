# Get default VPC
data "aws_vpc" "default" {
  default = true
}

# Get default subnets
data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# Security Group for RDS (allows connections from anywhere)
resource "aws_security_group" "rds" {
  name        = "jobping-rds-sg"
  description = "Security group for RDS PostgreSQL"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Allow from anywhere (for Lambda outside VPC)
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "jobping-rds-sg"
  }
}

# DB Subnet Group (required for RDS)
resource "aws_db_subnet_group" "main" {
  name       = "jobping-db-subnet"
  subnet_ids = data.aws_subnets.default.ids

  tags = {
    Name = "jobping-db-subnet"
  }
}

# RDS PostgreSQL Database
resource "aws_db_instance" "postgres" {
  identifier     = "jobping-db"
  engine         = "postgres"
  instance_class = "db.t3.micro"
  allocated_storage = 20
  
  db_name  = "jobscanner"
  username = "jobscanner"
  password = var.db_password
  
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids  = [aws_security_group.rds.id]
  publicly_accessible    = true
  skip_final_snapshot    = true
  
  tags = {
    Name = "jobping-db"
  }
}

output "db_endpoint" {
  value = aws_db_instance.postgres.endpoint
}

output "database_url" {
  value     = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
  sensitive = true
}

