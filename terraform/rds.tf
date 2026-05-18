resource "aws_db_instance" "postgres" {
  identifier        = "api-managementz-db"
  engine            = "postgres"
  engine_version    = "16"
  instance_class    = "db.t3.micro"
  allocated_storage = 20

  db_name  = "postgres"
  username = "api_managementz"
  password = var.db_password

  vpc_security_group_ids = [aws_security_group.api.id]
  publicly_accessible    = true
  skip_final_snapshot    = true

  tags = {
    Name = "api-managementz-db"
  }
}

output "rds_endpoint" {
  value = aws_db_instance.postgres.endpoint
}