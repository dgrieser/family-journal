FROM golang:1.27-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o server ./cmd/server

FROM alpine:3.24
WORKDIR /app
RUN adduser -D appuser
COPY --from=build /app/server ./server
COPY migrations ./migrations
RUN mkdir -p /app/uploads
RUN chown -R appuser:appuser /app/uploads && chmod -R u=rwX,go=rX /app/uploads
USER appuser
EXPOSE 8080
CMD ["/app/server"]
