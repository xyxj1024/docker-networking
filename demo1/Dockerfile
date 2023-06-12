FROM golang:latest AS builder

COPY main.go $GOPATH
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go

FROM ubuntu:latest

COPY --from=builder /go/main /bin
ENTRYPOINT [ "/bin/main" ]
CMD [ "9000" ]
EXPOSE 9000