FROM golang:1.15.1 as builder
WORKDIR /go/src/github.com/dddpaul/vscalebot
ADD . ./
RUN make build-alpine

FROM alpine:3.12
LABEL maintainer="Pavel Derendyaev <dddpaul@gmail.com>"
RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/* && \
    update-ca-certificates
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/vscalebot/bin/vscalebot .
#EXPOSE 8080

ENTRYPOINT ["./vscalebot"]
#CMD ["-port", ":8080"]
