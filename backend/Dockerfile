ARG IMAGE_REGISTRY_PREFIX=docker.m.daocloud.io/library/
FROM ${IMAGE_REGISTRY_PREFIX}golang:1.21-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY . .
COPY config.ini .
RUN go mod tidy && go build -o main .

ARG IMAGE_REGISTRY_PREFIX=docker.m.daocloud.io/library/
FROM ${IMAGE_REGISTRY_PREFIX}alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/config.ini .
EXPOSE 8080
CMD ["./main"]
