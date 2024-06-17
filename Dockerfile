FROM golang:alpine as builder

COPY . .

RUN go build -o /go/bin/app main.go

FROM scratch

FROM scratch
COPY --from=builder /go/bin/app /app

ENTRYPOINT ["/app"]
