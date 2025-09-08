# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/app /app/app
ENV ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/app/app"]

