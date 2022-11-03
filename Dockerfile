FROM golang:1.19.3 as builder

WORKDIR /build
COPY . /build/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/out/dashboard .



FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root
COPY --from=builder /build/out/dashboard .

CMD ["./dashboard"]
