# HSocks5

![Snapshot Build](https://github.com/Soontao/hsocks5/workflows/Snapshot%20Build/badge.svg)
![CircleCI](https://img.shields.io/circleci/build/github/Soontao/hsocks5)
[![codecov](https://codecov.io/gh/Soontao/hsocks5/branch/master/graph/badge.svg)](https://codecov.io/gh/Soontao/hsocks5)
[![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/thedockerimages/hsocks5)](https://hub.docker.com/repository/docker/thedockerimages/hsocks5)
[![Size](https://shields.beevelop.com/docker/image/image-size/thedockerimages/hsocks5/latest.svg?style=flat-square)](https://hub.docker.com/repository/docker/thedockerimages/hsocks5)

Provide HTTP Proxy based on Socks5 Proxy. 

This project is the `golang` version of the tool [http-proxy-to-socks](https://github.com/Soontao/http-proxy-to-socks), and much faster processing & less memory taking than it.

## Setup

just run this tool with `docker`

```bash
docker run --restart=always -d -p 18080:18080 -e SOCKS=192.168.3.88:10080 --name hsocks5 thedockerimages/hsocks5:latest
```

The `192.168.3.88:10080` is the socks5 server host and port.

The `18080` is the http proxy default port, you can use docker expose it as another port.

## Prometheus Metric 

`HSocks5` exposes prometheus metric endpoint at `/hsocks5/__/metric`