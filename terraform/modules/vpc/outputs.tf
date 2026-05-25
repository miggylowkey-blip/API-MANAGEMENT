output "vpc_id" {
  value = aws_vpc.this.id
}

output "public_subnet_ids" {
  value = [for subnet in aws_subnet.public : subnet.id]
}

output "private_subnet_ids" {
  value = [for subnet in aws_subnet.private : subnet.id]
}

output "public_subnet_arns" {
  value = [for subnet in aws_subnet.public : subnet.arn]
}

output "private_subnet_arns" {
  value = [for subnet in aws_subnet.private : subnet.arn]
}

output "public_route_table_ids" {
  value = [aws_route_table.public.id]
}

output "private_route_table_ids" {
  value = [for rt in aws_route_table.private : rt.id]
}
