FROM golang as builder

WORKDIR /app

ADD go.mod .
RUN go mod download

ADD . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o binary github.com/c16a/hermes/app

FROM alpine
WORKDIR /app

ENV CONFIG_FILE_PATH="/var/config.json"

COPY --from=builder /app/binary .
CMD /app/binary
