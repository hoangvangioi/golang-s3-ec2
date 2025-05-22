terraform {
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

# VPC and Security Group
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "app-vpc"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "main-igw"
  }
}

# Public Subnet
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.0.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet"
  }
}

# Route Table for Public Subnet
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "public-rt"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# Security Group for EC2
resource "aws_security_group" "ec2" {
  name        = "ec2-security-group"
  description = "Security group for EC2 instance"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "ec2-security-group"
  }
}

# Key Pair
resource "aws_key_pair" "app_key" {
  key_name   = "go-web-key"
  public_key = file("~/.ssh/go-web-key.pub")
}

# S3 Bucket
resource "aws_s3_bucket" "uploads" {
  bucket = var.s3_bucket_name

  tags = {
    Name = "file-uploads-bucket"
  }
}

resource "aws_s3_bucket_public_access_block" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "uploads" {
  bucket = aws_s3_bucket.uploads.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.uploads.arn}/*"
      }
    ]
  })
}

# EC2 Instance
resource "aws_instance" "app" {
  ami           = var.ec2_ami_id
  instance_type = var.ec2_instance_type
  subnet_id     = aws_subnet.public.id
  key_name      = aws_key_pair.app_key.key_name

  vpc_security_group_ids = [aws_security_group.ec2.id]

  user_data = <<-EOF
              #!/bin/bash
              # Update system
              dnf update -y
              
              # Install Go and Nginx
              dnf install -y golang nginx
              
              # Create app directory
              mkdir -p /app
              
              # Start and enable Nginx
              systemctl start nginx
              systemctl enable nginx
              EOF

  tags = {
    Name = "app-server"
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

# Outputs
output "ec2_public_ip" {
  value = aws_instance.app.public_ip
}

output "s3_bucket_name" {
  value = aws_s3_bucket.uploads.id
}

output "app_url" {
  value = "http://${aws_instance.app.public_ip}"
}
