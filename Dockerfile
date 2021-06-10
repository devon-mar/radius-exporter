FROM golang:1.15 as builder
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make

FROM scratch
COPY --from=builder /go/src/app/radius-exporter /radius-exporter
EXPOSE 9881
ENTRYPOINT ["/radius-exporter"]
