# Hermes
Hermes is a tiny message broker written in Go.

[![Go Workflow Status](https://github.com/c16a/hermes/workflows/Go/badge.svg)](https://github.com/c16a/hermes/workflows/Go/badge.svg)

[![CodeQL Workflow Status](https://github.com/c16a/hermes/workflows/CodeQL/badge.svg)](https://github.com/c16a/hermes/workflows/CodeQL/badge.svg)

[![Go Report Card](https://goreportcard.com/badge/github.com/c16a/hermes)](https://goreportcard.com/report/github.com/c16a/hermes)

The goals of the project are as below
- Easy to compile, and run
- Tiny footprint
- Extensible
- Adhering to standards

## Current features
This is in no way ready to be consumed. 
This is a project which arose out of my boredom during COVID-19, 
and general issues whilst working with other production ready brokers 
such as ActiveMQ, Solace, NATS etc.

### Simple text-based protocol
This was inspired by NATS, which is by far my favorite message broker.

```shell
$ telnet localhost 4000
Trying ::1...
Connected to localhost.
Escape character is '^]'.
PUB topic-1 Hello
OK
SUB client-1 group-1 my-topic
OK
```

Support for MQTT v5 is currently being worked on [feature/mqtt](https://github.com/c16a/hermes/tree/feature/mqtt). The custom protocol will soon be sunset in favor of MQTT.

### Supports client grouping
Multiple clients can subscribe, acting as a single unit, 
and the broker can randomly push the payload to just one of them.

## Planned features
The following are some features from the top of my head which I will work on
- Support for more transports such as WebSocket, gRPC, Rsocket(?) 
- Support for clustering
- Authentication & extensible middleware
- Message Persistence

## Contributing
Fork it, give it a spin, and let me know! 

