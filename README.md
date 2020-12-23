# Hermes
Hermes is a tiny message broker written in Go.

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

On retrospection, it now feels sensible to make this [MQTT-compatible](https://docs.oasis-open.org/mqtt/mqtt/v5.0/mqtt-v5.0.pdf). It is also dawning on me that the MQTT v5.0 spec is 137 pages long, and I'm too lazy to read through it. Parking it for the new year.

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

