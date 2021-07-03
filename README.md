# fbox - Files in a Box

![GitHub All Releases](https://img.shields.io/github/downloads/prologic/fbox/total)
![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/prologic/fbox)
![Docker Pulls](https://img.shields.io/docker/pulls/prologic/fbox)
![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/prologic/fbox)

![](https://github.com/prologic/fbox/workflows/Go/badge.svg)
![](https://github.com/prologic/fbox/workflows/ReviewDog/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/prologic/fbox)](https://goreportcard.com/report/prologic/fbox)
[![codebeat badge](https://codebeat.co/badges/15fba8a5-3044-4f40-936f-9e0f5d5d1fd9)](https://codebeat.co/projects/github-com-prologic-fbox-master)
[![GoDoc](https://godoc.org/github.com/prologic/fbox?status.svg)](https://godoc.org/github.com/prologic/fbox)

`fbox` is a distributed file system written in [Go](https://golang.org)
that has the the following features:

**Current Features**:

- Single portable binary
- Simple to setup and maintain
- A Web Interface with Drag 'n Drop
- Data redundancy with erasure coding and sharding

**Planned Features:**

- POSIX compatible FUSE interface
- S3 Object Storage interface
- Docker Volume Driver

There is also a publicly (_freely_) available demo instance available at:

- https://files.mills.io/

> __NOTE:__ I, [James Mills](https://github.com/prologic), run this instance
on pretty cheap hardware on a limited budget. Please use it fairly so everyone
can enjoy using it equally!

> **[Sponsor](#Sponsor)** this project to support the development of new features,
> improving existings ones and fix bugs!

![Screenshot](/docs/screenshot.png)

Table of Contents
=================

* [fbox \- Files in a Box](#fbox---files-in-a-box)
* [Table of Contents](#table-of-contents)
  * [Why?](#why)
  * [Getting Started](#getting-started)
    * [Install from Releases](#install-from-releases)
    * [Install from Homebrew](#install-from-homebrew)
    * [Install from Source](#install-from-source)
  * [Usage](#usage)
  * [Prepare storage](#prepare-storage)
  * [Start a master node](#start-a-master-node)
  * [Start some data nodes](#start-some-data-nodes)
  * [License](#license)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

## Why?

`fbox` was written primarily and firstly as en education exercise in distributed
file systems and a lot of time add effort went into understanding how many
distributed file systems in the wild are designed.

Most other distributed file systems are either complex and hard to set up
and maintain or come with really expensive license feeds. Therefore
`fbox` is also an attempt at designing and implementing a distributed
file system that is free and open source with the reliability, scale
and durability of other distributed file systems.

## Getting Started

### Install from Releases

You can install `fbox` by simply downloading the latest version from the
[Release](https://github.com/prologic/fbox/releases) page for your platform
and placing the binary in your `$PATH`.

For convenience you can run one of the following shell pipelines which will
download and install the latest release binary into `/usr/local/bin`
(_modify to suit_):

For Linux x86_64:

```console
curl -s https://api.github.com/repos/prologic/fbox/releases/latest | grep browser_download_url | grep Linux_x86_64 | cut -d '"' -f 4 | wget -q -O - -i - | tar -zxv fbox && mv fbox /usr/local/bin/fbox
```

For macOS x86_64:

```console
curl -s https://api.github.com/repos/prologic/fbox/releases/latest | grep browser_download_url | grep Darwin_x86_64 | cut -d '"' -f 4 | wget -q -O - -i - | tar -zxv fbox && mv fbox /usr/local/bin/fbox
```

### Install from Homebrew

On macOS you can install `fbox` using [Homebrew](https://brew.sh):

```#!console
brew tap prologic/fbox
brew install fbox
```

### Install from Source

To install `fbox` from source you can run `go get` directly if you have a [Go](https://golang.org) environment setup:

```#!console
go get github.com/prologic/fbox
```

> __NOTE:__ Be sure to have `$GOBIN` (_if not empty_) or your `$GOPATH/bin`
>           in your `$PATH`.
>           See [Compile and install packages and dependencies](https://golang.org/cmd/go/#hdr-Compile_and_install_packages_and_dependencies)

Or grab the source code and build:

```#!console
git clone https://github.com/prologic/fbox.git
cd fbox
make build
```

And optionally run `make install` to place the binary `fbox` in your `$GOBIN`
or `$GOPATH/bin` (_again see note above_).

## Usage

## Prepare storage

```#!console
mkdir data1 data2 data3
```

## Start a master node

```#!console
fbox -a 127.0.0.1:8000 -b :8000 -d ./data1
```

## Start some data nodes

```#!console
fbox -a 127.0.0.1:8001 -b :8001 -d ./data2 -m http://127.0.0.1:8000
fbox -a 127.0.0.1:8002 -b :8002 -d ./data3 -m http://127.0.0.1:8000
```

This will set up a 3-node cluster on your local machine (_for testing only_).

You can open the Web UI by navigating to:

http://127.0.0.1:8000/ui/

You can also use the command-line client `fbox` itself; See usage:

```#!console
$ ./fbox --help
Usage: fbox [options] [command [arguments]]

fbox is a simple distributed file system...

Valid commands:
 - cat <name> -- Downloads the given file given by <name> to stdout
 - put <name> -- Upload the given file given by <name>

Valid options:
  -a, --advertise-addr string   [interface]:port to advertise
  -b, --bind string             [interface]:port to bind to (default "0.0.0.0:8000")
  -s, --data-shards int         no. of data shards (default 3)
  -D, --debug                   enable debug logging
  -d, --dir string              path to store data in (default "./data")
  -m, --master string           address:port of master
  -p, --parity-shards int       no. of parity shards (default 1)
  -V, --version                 display version information and exit
pflag: help requested
```

For example to store a file:

```#!console
echo "Hello World" > hello.txt
fbox -m http://127.0.0.1:8000 put hello.txt
```

And to retrieve the file:

```#!console
fbox -m http://127.0.0.1:8000 cat hello.txt
```

## Production Deployments

### Docker Swarm

You can deploy `fbox` to a [Docker Swarm](https://docs.docker.com/engine/swarm/)
cluster by utilising the provided `fbox.yml` Docker Stack. This also depends on
and uses the [Traefik](https://docs.traefik.io/) ingress load balancer so you must
also have that configured and running in your cluster appropriately.

```console
export DOMAIN=files.yourdomain.tld
docker stack deploy -c fbox.yml fbox
```

## License

`fbox` is licensed under the terms of the [MIT License](/LICENSE)
