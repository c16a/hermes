FROM alpine:edge as builder

RUN apk update
RUN apk upgrade
RUN apk add --update go gcc g++

WORKDIR /app

ADD go.mod .
RUN go mod download

ADD . .
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o binary github.com/c16a/hermes/app

FROM alpine
WORKDIR /app

ENV CONFIG_FILE_PATH="/var/config.json"

COPY --from=builder /app/binary .
CMD ["/app/binary"]
