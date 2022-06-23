FROM gcr.io/distroless/static-debian11

COPY bin/server /usr/local/bin/server

EXPOSE 9090
EXPOSE 8080
EXPOSE 50051

ENTRYPOINT ["server"]

#FROM golang:1.18 AS builder
#
#WORKDIR /workspace
#
#COPY go.mod go.mod
#COPY go.sum go.sum
#COPY pkg/ pkg/
#COPY cmd/ cmd/
#COPY proto/ proto/
#
#RUN go mod tidy
#RUN go build cmd -o build

#FROM alpine:latest
#WORKDIR /
#COPY --from=builder workspace/bin/server /usr/local/bin
#USER 1000:1000
#ENTRYPOINT ["server"]