# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS build

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache wget ca-certificates && adduser -D -u 10001 app

COPY --from=build /out/api /api
COPY api.yaml /api.yaml

EXPOSE 8080

USER app
ENTRYPOINT ["/api"]
