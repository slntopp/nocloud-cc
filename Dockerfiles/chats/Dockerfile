ARG IMAGE_TAG=latest
ARG ARCH=amd64
FROM ghcr.io/slntopp/nocloud-cc/base:${IMAGE_TAG} as builder

RUN CGO_ENABLED=0 GOARCH=${ARCH} go build -ldflags="-s -w" -buildvcs=false ./cmd/chat
RUN upx ./chat

RUN adduser -SD nocloud

FROM scratch
WORKDIR /
COPY --from=builder  /go/src/github.com/slntopp/nocloud-cc/chat /chat
COPY --from=builder /etc/passwd /etc/passwd

USER nocloud
EXPOSE 8000
LABEL org.opencontainers.image.source https://github.com/slntopp/nocloud-cc
LABEL nocloud.update "true"

ENTRYPOINT ["/chat"]
