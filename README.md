# Hermes

Hermes is a tiny MQTT compatible broker written in Go.

[![Go Workflow Status](https://github.com/c16a/hermes/workflows/Go/badge.svg)](https://github.com/c16a/hermes/workflows/Go/badge.svg)

[![CodeQL Workflow Status](https://github.com/c16a/hermes/workflows/CodeQL/badge.svg)](https://github.com/c16a/hermes/workflows/CodeQL/badge.svg)

[![Go Report Card](https://goreportcard.com/badge/github.com/c16a/hermes)](https://goreportcard.com/report/github.com/c16a/hermes)

[![Total alerts](https://img.shields.io/lgtm/alerts/g/c16a/hermes.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/c16a/hermes/alerts/)

The goals of the project are as below

- Easy to compile, and run
- Tiny footprint
- Extensible
- Adhering to standards

## Current features

This is in no way ready to be consumed. This is a project which arose out of my boredom during COVID-19, and general
issues whilst working with other production ready brokers such as ActiveMQ, Solace, NATS etc.

- [x] CONNECT
- [x] PUBLISH, PUBACK
- [x] SUBSCRIBE, SUBACK
- [x] DISCONNECT
- [x] Persistent sessions
- [x] QoS 2 support
- [x] Offline messages
- [ ] Wildcard subscriptions
- [ ] Shared Subscriptions
- [ ] Extended authentication
- [ ] MQTT over WebSocket
- [ ] Clustering

## Usage

Any compatible MQTT client library can be used to interact with the broker

- Java ([eclipse/paho.mqtt.java](https://github.com/eclipse/paho.mqtt.java))
- Go ([eclipse/paho.golang](https://github.com/eclipse/paho.golang))
- Other clients can be found [here](https://github.com/eclipse?q=paho&type=&language=)

## Planned features

The following are some features from the top of my head which I will work on

- Support for more transports such as WebSocket, gRPC, Rsocket(?)
- Support for clustering
- Authentication & extensible middleware
- Message Persistence

## Contributing

Fork it, give it a spin, and let me know! 

