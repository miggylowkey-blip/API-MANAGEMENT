# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# Run stage
FROM alpine:3.19

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./api"]