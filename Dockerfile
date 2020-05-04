FROM golang:latest AS builder
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

FROM debian:stable-slim
COPY --from=builder /go/bin/app .
COPY public /public
EXPOSE 1337
ENTRYPOINT ["/app"]