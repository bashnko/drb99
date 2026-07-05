FROM golang:1.26 AS builder

WORKDIR /app
COPY . .

ENV CGO_ENABLED=0

RUN go build -o drb99 cmd/drb99/main.go 

FROM alpine
COPY --from=builder /app/drb99 .

CMD ["./drb99"]
