ARG GO_IMAGE=golang:1.26.2-alpine
ARG DOCKER_CLI_IMAGE=docker:28-cli
ARG API_RUNTIME_IMAGE=alpine:3.21

FROM ${GO_IMAGE} AS builder

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/senspace main.go

FROM ${DOCKER_CLI_IMAGE} AS dockercli

FROM ${API_RUNTIME_IMAGE}

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/senspace /app/senspace
COPY --from=builder /src/asset /app/asset
COPY --from=dockercli /usr/local/bin/docker /usr/local/bin/docker

ENV SENSPACE_ENV=uat

EXPOSE 9010

CMD ["/app/senspace"]
