KT-Connect
===========

![Go](https://github.com/alibaba/kt-connect/workflows/Go/badge.svg)
[![Build Status](https://travis-ci.org/alibaba/kt-connect.svg?branch=master)](https://travis-ci.org/alibaba/kt-connect)
[![Go Report Card](https://goreportcard.com/badge/github.com/alibaba/kt-connect)](https://goreportcard.com/report/github.com/alibaba/kt-connect)
[![Test Coverage](https://api.codeclimate.com/v1/badges/eb13b3946784bd7c67cc/test_coverage)](https://codeclimate.com/github/alibaba/kt-connect/test_coverage)
[![Maintainability](https://api.codeclimate.com/v1/badges/eb13b3946784bd7c67cc/maintainability)](https://codeclimate.com/github/alibaba/kt-connect/maintainability)
[![Release](https://img.shields.io/github/release/alibaba/kt-connect.svg?style=flat-square)](https://img.shields.io/github/release/alibaba/kt-connect.svg?style=flat-square)
![License](https://img.shields.io/github/license/alibaba/kt-connect.svg)

English | [简体中文](./README_CN.md)

KtConnect ("Kt" is short for "Kubernetes Toolkit") is a utility tool to help you work with Kubernetes dev environment more efficiently.

![Arch](./docs/media/arch.png)

## ✅ Features

* `Connect`: Directly Access a remote Kubernetes cluster. KtConnect use ssh-vpn or socks-proxy to access remote Kubernetes cluster networks.
* `Exchange`: Developer can exchange the workload to redirect the requests to a local app.
* `Mesh`: You can create a mesh version service in local host, and redirect specified workload requests to your local.
* `Preview`: Expose a local running app to Kubernetes cluster as a common service, all requests to that service are redirect to local app.

## 🚀 QuickStart

You can download and install the `ktctl` from [Downloads And Install](docs/en-us/guide/downloads.md)

Read the [Quick Start Guide](docs/en-us/guide/quickstart.md) for more about this tool.

## 🔨 Building from Source

This project requires **Go 1.18** to compile. Building with newer Go versions (e.g. 1.24) will fail due to incompatible dependencies (`gvisor` and `go-shadowsocks2`).

If you only have a newer Go installed, use the official Go toolchain manager to install 1.18 side-by-side:

```bash
go install golang.org/dl/go1.18@latest
go1.18 download

# Build with Go 1.18
GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct go1.18 build -o ktctl ./cmd/ktctl/
```

The `go1.18` binary is installed to `~/go/bin/go1.18` and does not affect your system Go.

## 💡 Ask For Help

Please feel free to raise an [issue](https://github.com/alibaba/kt-connect/issues) if anything sucks, or go ahead to contact us with DingTalk（钉钉）:

<img src="https://img.alicdn.com/imgextra/i4/O1CN01sTW3D61NzAFgUCNqz_!!6000000001640-0-tps-573-657.jpg" width="50%"></img>

