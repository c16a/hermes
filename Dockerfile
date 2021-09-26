FROM docker.io/golang:1.17 as builder

WORKDIR /app

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . .

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o binary github.com/c16a/hermes/app

FROM scratch
WORKDIR /app

ENV CONFIG_FILE_PATH="/var/config.json"

COPY --from=builder /app/binary .
CMD ["/app/binary"]
