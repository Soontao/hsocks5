# HSocks5

![Snapshot Build](https://github.com/Soontao/hsocks5/workflows/Snapshot%20Build/badge.svg)
![CircleCI](https://img.shields.io/circleci/build/github/Soontao/hsocks5)
[![codecov](https://codecov.io/gh/Soontao/hsocks5/branch/master/graph/badge.svg)](https://codecov.io/gh/Soontao/hsocks5)
[![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/thedockerimages/hsocks5)](https://hub.docker.com/repository/docker/thedockerimages/hsocks5)
[![Size](https://shields.beevelop.com/docker/image/image-size/thedockerimages/hsocks5/latest.svg?style=flat-square)](https://hub.docker.com/repository/docker/thedockerimages/hsocks5)

Provide HTTP Proxy based on Socks5 Proxy. 

This project is the `golang` version of the tool [http-proxy-to-socks](https://github.com/Soontao/http-proxy-to-socks), with much faster processing & less memory taking.

## Why? 

Most `operation systems` (like `windows`, `android` & `iOS`) only support `HTTP Proxy` without other tools, if users want to use a `socks5` proxy, users must install some app on device, but sometimes there are some limitation on the devices. This project can `transform` `socks5` proxy into `http` proxy, so that make all devices can connect with the `socks5` proxy.

## Setup with binary

Download released binaries from [here](https://github.com/Soontao/hsocks5/releases). (you should download the correct binary for your platform)

And run with 

```bash
./hsocks5 --socks 192.168.3.88:10080 start
```

## Setup with docker

Run this tool with `docker`

```bash
docker run --restart=always -d -p 18080:18080 -e SOCKS=192.168.3.88:10080 --name hsocks5 thedockerimages/hsocks5:latest
```

The `192.168.3.88:10080` is the socks5 server host and port.

The `18080` is the http proxy default port, you can use docker expose it as another port.

## Options

```bash
./hsocks5 --help

NAME:
   HSocks5 - provide HTTP Proxy with Socks5

USAGE:
   hsocks5 [global options] command [command options] [arguments...]

VERSION:
   SNAPSHOT

COMMANDS:
   start    program entry
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --addr value, -a value   HTTP proxy listen address (default: ":18080") [%ADDR%]
   --socks value, -s value  Socks5 service url, format: hostname:port, 192.168.1.1:18080 [%SOCKS%]
   --help, -h               show help
   --version, -v            print the version
```

## Prometheus metric 

`HSocks5` exposes prometheus metric endpoint at `/hsocks5/__/metric`

## [CHANGELOG](./CHANGELOG.md)