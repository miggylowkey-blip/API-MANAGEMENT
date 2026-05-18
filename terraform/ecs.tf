resource "aws_ecs_cluster" "main" {
  name = "api-managementz-cluster"
}

resource "aws_iam_role" "ecs_task_execution" {
  name = "api-managementz-ecs-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_ecs_task_definition" "api" {
  family                   = "api-managementz-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn

  container_definitions = jsonencode([{
    name  = "api-managementz"
    image = "${aws_ecr_repository.api.repository_url}:latest"
    portMappings = [{
      containerPort = 8080
      protocol      = "tcp"
    }]
    environment = [
      { name = "PORT",             value = "8080" },
      { name = "DB_HOST",          value = split(":", aws_db_instance.postgres.endpoint)[0] },
      { name = "DB_USER",          value = "api_managementz" },
      { name = "DB_PASSWORD",      value = var.db_password },
      { name = "DB_NAME",          value = "postgres" },
      { name = "DB_SSLMODE",       value = "require" },
      { name = "AUTH_SECRET",      value = var.auth_secret },
      { name = "RATE_LIMIT_RPS",   value = "5" },
      { name = "RATE_LIMIT_BURST", value = "10" },
      { name = "REDIS_URL",        value = "redis://localhost:6379" }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = "/ecs/api-managementzz-task"
        awslogs-region        = var.region
        awslogs-stream-prefix = "ecs"
      }
    }
  }])
}

resource "aws_cloudwatch_log_group" "api" {
  name              = "/ecs/api-managementzz-task"
  retention_in_days = 7
}

resource "aws_ecs_service" "api" {
  name            = "api-managementzz-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.default.ids
    security_groups  = [aws_security_group.api.id]
    assign_public_ip = true
  }
}