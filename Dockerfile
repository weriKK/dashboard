FROM golang:1.19.3 as builder

WORKDIR /build
COPY . /build/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/out/dashboard-backend ./cmd/dashboard-backend



FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /home
COPY --from=builder /build/out/dashboard-backend .
COPY config.yaml .

CMD ["./dashboard-backend"]
