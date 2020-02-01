FROM golang:alpine
COPY src /go/src
WORKDIR /go/src
RUN go build *.go

FROM alpine
COPY --from=0 /go/src/main /main
COPY --from=0 /go/src/config.yaml /config.yaml
CMD ["./main"]
