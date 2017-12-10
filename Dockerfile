FROM golang:1.8.3 as builder
WORKDIR /go/src/github.com/dddpaul/vscalebot
ADD . ./
RUN make build-alpine

FROM alpine:latest
RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/* && \
    update-ca-certificates
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/vscalebot/bin/vscalebot .
#EXPOSE 8080

ENTRYPOINT ["./vscalebot"]
#CMD ["-port", ":8080"]
