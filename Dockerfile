FROM golang:1.19.3-alpine3.17 as builder

ARG GOLANG_NAMESPACE="github.com/mohemohe/mastoguard"
ENV GOLANG_NAMESPACE="$GOLANG_NAMESPACE"

RUN apk --no-cache add alpine-sdk coreutils make tzdata
RUN cp -f /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
WORKDIR /go/src/$GOLANG_NAMESPACE
ADD ./go.* /go/src/$GOLANG_NAMESPACE/
ENV GO111MODULE=on
RUN go mod download
ADD . /go/src/$GOLANG_NAMESPACE/
RUN make build
RUN mkdir -p /mastoguard
RUN mv /go/src/$GOLANG_NAMESPACE/mastoguard /mastoguard/

# ====================================================================================

FROM alpine

RUN apk --no-cache add ca-certificates
COPY --from=builder /etc/localtime /etc/localtime
COPY --from=builder /mastoguard /mastoguard

WORKDIR /mastoguard
CMD ["/mastoguard/mastoguard"]