FROM --platform=linux/amd64 golang:alpine3.18 AS builder

WORKDIR /app

ENV GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags "-s"

FROM alpine:3.18

EXPOSE 53/udp
COPY --from=builder /app/docker-demo6 /usr/local/bin/
CMD ["/usr/local/bin/docker-demo6", \
     "--ratelimit", "2", \
     "--domain", "swm.example.com", \
     "--domain", "anotherswm.example.com"]