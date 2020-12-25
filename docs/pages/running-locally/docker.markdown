---
layout: page
title: Using Docker
nav_order: 2
has_children: false
parent: Running locally
permalink: /running-locally/docker
---

Hermes uses a multi-stage Docker build for a minimal runtime.

```shell
git clone https://github.com/c16a/hermes
cd hermes
docker build -t hermes-image .
```
