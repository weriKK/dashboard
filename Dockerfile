FROM golang:1.11.1 as builder

RUN go get -d -v github.com/weriKK/dashboard
WORKDIR /go/src/github.com/weriKK/dashboard
COPY . /go/src/github.com/weriKK/dashboard/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dashboard .



FROM alpine:latest

RUN apk --no-cache add ca-certificates rsyslog
COPY ./rsyslog.conf /etc/rsyslog.conf

WORKDIR /root
COPY --from=builder /go/src/github.com/weriKK/dashboard/dashboard .
COPY ./dockerStartup.sh /root/dockerStartup.sh

CMD ["/bin/sh", "/root/dockerStartup.sh"]
