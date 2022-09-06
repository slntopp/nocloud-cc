ARG ARCH=amd64

FROM golang:1.19-alpine AS builder

RUN apk add upx
# Add CA Certificates for those services communicating with outerworld
RUN apk add -U --no-cache ca-certificates

WORKDIR /go/src/github.com/slntopp/nocloud-cc
COPY go.mod go.sum ./
RUN go mod download

COPY pkg pkg
COPY cmd cmd

LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud-cc
LABEL nocloud.update "true"
