# backend/Dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server  ./cmd/app
RUN CGO_ENABLED=0 go build -o migrate ./cmd/migrate

FROM alpine:3.19
WORKDIR /app
# Both binaries ship in the image. migrations are embedded (embed.FS, EL-056),
# so no migrations/ directory is copied. Railway's preDeployCommand runs
# ./migrate up before the new release serves traffic (railway.toml).
COPY --from=builder /app/server  .
COPY --from=builder /app/migrate .
EXPOSE 8080
CMD ["./server"]
