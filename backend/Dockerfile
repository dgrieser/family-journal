FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/api

FROM alpine:latest

WORKDIR /app

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN mkdir -p uploads && chown -R appuser:appgroup /app

COPY --from=builder --chown=appuser:appgroup /app/main .

# Switch to the non-root user
USER appuser

EXPOSE 8080

CMD ["./main"]
