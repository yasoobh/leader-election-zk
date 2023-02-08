FROM golang:1.19.5-buster AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o dist-logs .

FROM arm64v8/ubuntu:23.04

RUN ls /
COPY --from=builder /app/dist-logs /dist-logs
RUN chmod +x /dist-logs

CMD ["./dist-logs"]
