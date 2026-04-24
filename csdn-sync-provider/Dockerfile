ARG IMAGE_REGISTRY_PREFIX=docker.m.daocloud.io/library/
FROM ${IMAGE_REGISTRY_PREFIX}golang:1.21-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY . .
RUN go mod tidy && go build -o csdn-sync-provider .

ARG IMAGE_REGISTRY_PREFIX=docker.m.daocloud.io/library/
FROM ${IMAGE_REGISTRY_PREFIX}alpine:latest
WORKDIR /app
COPY --from=builder /app/csdn-sync-provider .
EXPOSE 8091
CMD ["./csdn-sync-provider"]
