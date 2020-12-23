FROM golang as builder

WORKDIR /app

ADD go.mod .
RUN go mod download

ADD . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o binary_amd64 github.com/c16a/hermes/app

FROM scratch
WORKDIR /app
COPY --from=builder /app/binary_amd64 .
CMD ["/app/binary_amd64"]
