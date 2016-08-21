# quantum
[![Build Status](https://travis-ci.org/Supernomad/quantum.svg?branch=develop)](https://travis-ci.org/Supernomad/quantum) [![Coverage Status](https://coveralls.io/repos/github/Supernomad/quantum/badge.svg?branch=develop)](https://coveralls.io/github/Supernomad/quantum?branch=develop) [![Go Report Card](https://goreportcard.com/badge/github.com/Supernomad/quantum)](https://goreportcard.com/report/github.com/Supernomad/quantum) [![GoDoc](https://godoc.org/github.com/Supernomad/quantum?status.png)](https://godoc.org/github.com/Supernomad/quantum)

`quantum` is a software defined network device written in go with global networking, security, and auto-configuration at its heart. It leverages the latest distributed data stores and state of the art encryption to offer fully secured end to end global networking over a single cohesive network.

- Encrypted with [AES-256-GCM](http://crypto.stackexchange.com/questions/17999/aes256-gcm-can-someone-explain-how-to-use-it-securely-ruby).
  - Ensuring both confidentiality of all network traffic but also authentication of both the recipient and sender.
- Secret generation using [ECDHE](https://en.wikipedia.org/wiki/Elliptic_curve_Diffie%E2%80%93Hellman) with [Curve25519](https://en.wikipedia.org/wiki/Curve25519).
  - Ensuring secret generation/transmission is always as secure as possible.
- Fully peer to peer communication.
  - Minimizing bottlenecks and maximizing performance.
- Full DHCP and static ip address designation.
  - Internal private ip network must be ipv4.
  - External public ip network can be any combination of ipv4/ipv6.
- Lightweight and designed to run on systems with limited available resources.
- Designed with global distributions in mind.
  - Ability to scale to thousands of nodes spanning any geographic region.

##### Supported Data stores
- Consul
- Etcd

##### Supported OS's
- Linux

###### Soon to be supported:
- BSD
- Darwin

### Running `quantum`


### Development
Currently `quantum` development is entirely in go and utilizes a few BASH scripts to facilitate builds and setup. Development has been mostly done on ubuntu server 14.04, however any recent linux distribution with the following dependencies should be sufficient to develop `quantum`.

#### Development Dependencies
- bash
- tun kernel module must be enabled
  - please see your distributions information on how to enable it.
- docker
- docker-compose
- go 1.6

#### Getting started
To get started developing `quantum`, run the following shell commands to get your environment configured and running.

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
# Get build dependencies
$ go get -t -v ./...
$ go get golang.org/x/tools/cmd/cover
$ go get github.com/mattn/goveralls
$ go get github.com/golang/lint/golint
$ go get github.com/GeertJohan/fgt
# Run a build of quantum which will ensure your system is indeed up to date.
$ bin/build.sh
# Build the container to run quantum in.
$ docker-compose build
# Start up etcd
$ docker-compose up -d etcd
# Start up quantum
$ docker-compose up -d quantum0
$ docker-compose up -d quantum1
$ docker-compose up -d quantum2
# Check on the status of the different quantum instances
$ docker-compose logs quantum0 quantum1 quantum2
```
After running the above you will have a single etcd instance and three quantum instances running. The three quantum instances are configured to run a quantum network `10.9.0.0/16`, with `quantum0` having a statically defined private ip `10.9.0.1` and `quantum1`/`quantum2` having DHCP defined private ip addresses.

> A note about TLS, there are included TLS certificates in this repo that should not be for any reason reused for anything but testing quantum in a lab setting. There is an included script `generate-etcd-certs` that was used to create the certs in this repo if you are interested in reproducing them.

#### Configuration
`quantum` can be configured in any combination of four ways, cli arguments, environment variables, configuration file entries, and finally falling back to defaults. Regardless of which way `quantum` is configured all of the configuration options are available.

The precedence of the configuration methods is as follows:

- Cli arguments override their corresponding environment variables, configuration file entries, and defaults
- Environment variables override their corresponding configuration file entries and defaults
- Configuration file entries override their corresponding defaults

The naming of the environment variables and configuration file entries are all based on the cli argument names:

- Environment variables replace the prefixed `-` or `--` with `QUANTUM_`, replace any internal `-` with `_`, and convert the case to all upper for each cli argument.
  - Example: `conf-file` converts to `QUANTUM_CONF_FILE` and `tls-skip-verify` converts to `QUANTUM_TLS_SKIP_VERIFY`.
- Configuration file entires take the form of a flat `yaml` or `json` object, and drop the prefiexed `-` or `--` from each cli argument.
  - Example files:

    ``` yaml
    ---
    conf-file: "/etc/quantum/quantum.conf"
    public-ip: 1.1.1.1
    ```

    ``` json
    {
        "conf-file": "/etc/quantum/quantum.conf",
        "public-ip": "1.1.1.1"
    }
    ```


Run `quantum --help` for a current list of configuration options or see the [wiki](https://github.com/Supernomad/quantum/wiki/Configuration).

#### Testing
To run basic unit testing and builds run:

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ bin/build.sh
```

To do bandwidth based testing the `quantum` containers all have iperf3 installed. For example to test how much through put `quantum0` can handle from both `quantum1`/`quantum2`:

``` shell
# Assuming the three development quantum instances are started
# Start three shells

# In first shell start iperf3 server
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum0 iperf3 -s -f M

# In second shell start iperf3 client in quantum1
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum1 iperf3 -c 10.9.0.1 -P 2 -t 50

# In third shell start iperf3 client in quantum2
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum2 iperf3 -c 10.9.0.1 -P 2 -t 50
```

### Contributing
Contributions are definitely welcome, if you are looking for something to contribute check out the current [road map](https://github.com/Supernomad/quantum/milestones) and grab an open issue.

Workflow:

- Fork `quantum` develop
- Make your changes
- Open a PR against `quantum` develop
- Get your PR merged
- Rinse and Repeat

There are a few rules:

- All travis builds must successfully complete before a PR will be considered.
  - Changes to travis to get builds working are ok, if they are within reason.
- The `bin/build.sh` script must be run before the PR is open.
- Documentation is added for new user facing functionality

> An aside any PR can be closed with or without explanation.
