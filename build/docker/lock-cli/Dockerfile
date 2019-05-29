FROM golang:alpine as builder 

RUN apk --no-cache add ca-certificates

WORKDIR /go/src/github.com/kramergroup-workflows/lock-agent
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/lockcli cmd/lockcli/main.o

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/lockcli /lockcli

ENTRYPOINT [ "/lockcli" ]