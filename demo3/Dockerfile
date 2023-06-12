FROM golang:latest

COPY http_server.go .

ENV GO111MODULE=off

RUN go get golang.org/x/sys/unix

CMD ["go", "run", "http_server.go"]