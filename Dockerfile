FROM golang:alpine3.17 as builder
RUN mkdir "/src"
ADD . /src/
WORKDIR /src
RUN go build -ldflags "-s -w -X main.version=$(cat VERSION)" -o kontroller

FROM alpine
COPY --from=builder /src/kontroller /app/kontroller
WORKDIR /app
ENTRYPOINT ["/app/kontroller"]