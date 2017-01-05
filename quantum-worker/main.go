// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package main

import (
	"errors"
	"net/rpc"
	"os"

	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/device"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/workers"
)

func handleError(log *common.Logger, err error) {
	if err != nil {
		log.Error.Println(err.Error())
		os.Exit(1)
	}
}

func main() {
	log := common.NewLogger(common.DebugLogger)

	client, err := rpc.DialHTTP("unix", os.Args[1])
	if err != nil {
		handleError(log, errors.New("error reaching master process via rpc: "+err.Error()))
	}

	var cfg common.Config
	err = client.Call("ConfigServer.Sync", 0, &cfg)
	if err != nil {
		handleError(log, errors.New("error synchronizing config with master process via rpc: "+err.Error()))
	}

	dev, err := device.New(device.TUNDevice, &cfg)
	handleError(log, err)

	sock, err := socket.New(socket.UDPSocket, &cfg)
	handleError(log, err)

	outgoing := workers.NewOutgoing(log, &cfg, dev, sock)
	incoming := workers.NewIncoming(log, &cfg, dev, sock)

	incoming.Start(0)
	outgoing.Start(0)

	fds := make([]int, cfg.NumWorkers*2)
	copy(fds[0:cfg.NumWorkers], dev.Queues())
	copy(fds[cfg.NumWorkers:cfg.NumWorkers*2], sock.Queues())

	signaler := common.NewSignaler(log, &cfg, fds, map[string]string{common.RealDeviceNameEnv: dev.Name()})

	log.Info.Printf("[MAIN] Listening on TUN device:  %s", dev.Name())
	log.Info.Printf("[MAIN] TUN network space:        %s", cfg.NetworkConfig.Network)
	log.Info.Printf("[MAIN] TUN private IP address:   %s", cfg.PrivateIP)
	log.Info.Printf("[MAIN] TUN public IPv4 address:  %s", cfg.PublicIPv4)
	log.Info.Printf("[MAIN] TUN public IPv6 address:  %s", cfg.PublicIPv6)
	log.Info.Printf("[MAIN] Listening on UDP port:    %d", cfg.ListenPort)

	err = signaler.Wait(true)
	handleError(log, err)

	incoming.Stop()
	outgoing.Stop()

	err = sock.Close()
	handleError(log, err)

	err = dev.Close()
	handleError(log, err)
}
