variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "ap-southeast-1" # Singapore
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for file uploads"
  type        = string
  default     = "file-uploads-2024"
}

variable "ec2_instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

variable "ec2_ami_id" {
  description = "AMI ID for EC2 instance"
  type        = string
  default     = "ami-0afc7fe9be84307e4" # Amazon Linux 2023 tại Singapore
}
