FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o server ./cmd/server

FROM alpine:3.19
WORKDIR /app
RUN adduser -D appuser
COPY --from=build /app/server ./server
COPY migrations ./migrations
RUN mkdir -p /app/uploads && chown -R appuser:appuser /app/uploads
USER appuser
EXPOSE 8080
CMD ["/app/server"]
