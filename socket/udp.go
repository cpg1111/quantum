// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package socket

import (
	"errors"
	"syscall"

	"github.com/Supernomad/quantum/common"
)

// UDP socket struct for managing a multi-queue udp socket.
type UDP struct {
	queues []int
	cfg    *common.Config
}

// Close the UDP socket and remove associated network configuration.
func (sock *UDP) Close() error {
	for i := 0; i < len(sock.queues); i++ {
		if err := syscall.Close(sock.queues[i]); err != nil {
			return errors.New("error closing the socket queues: " + err.Error())
		}
	}
	return nil
}

// Queues will return the underlying UDP socket file descriptors.
func (sock *UDP) Queues() []int {
	return sock.queues
}

// Read a packet off the specified UDP socket queue and return a *common.Payload representation of the packet.
func (sock *UDP) Read(buf []byte, queue int) (*common.Payload, bool) {
	n, _, err := syscall.Recvfrom(sock.queues[queue], buf, 0)
	if err != nil {
		return nil, false
	}
	return common.NewSockPayload(buf, n), true
}

// Write a *common.Payload to the specified UDP socket queue.
func (sock *UDP) Write(payload *common.Payload, queue int) bool {
	err := syscall.Sendto(sock.queues[queue], payload.Raw[:payload.Length], 0, payload.Sockaddr)
	if err != nil {
		return false
	}
	return true
}

func newUDP(cfg *common.Config) (*UDP, error) {
	queues := make([]int, cfg.NumWorkers)
	sock := &UDP{queues: queues, cfg: cfg}
	for i := 0; i < sock.cfg.NumWorkers; i++ {
		var queue int
		var err error

		if !sock.cfg.ReuseFDS {
			queue, err = createUDP(sock.cfg.IsIPv6Enabled)
			if err != nil {
				return sock, errors.New("error creating the UDP socket: " + err.Error())
			}
			var la syscall.Sockaddr
			if sock.cfg.IsIPv6Enabled {
				sa := &syscall.SockaddrInet6{Port: cfg.ListenPort}
				copy(sa.Addr[:], cfg.ListenAddr.To16()[:])
				la = sa
			} else {
				sa := &syscall.SockaddrInet4{Port: cfg.ListenPort}
				copy(sa.Addr[:], cfg.ListenAddr.To4()[:])
				la = sa
			}
			err = initUDP(queue, la)
			if err != nil {
				return sock, err
			}
		} else {
			queue = 3 + sock.cfg.NumWorkers + i
		}
		sock.queues[i] = queue
	}
	return sock, nil
}

func initUDP(queue int, sa syscall.Sockaddr) error {
	err := syscall.SetsockoptInt(queue, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return errors.New("error setting the UDP socket parameters: " + err.Error())
	}

	err = syscall.Bind(queue, sa)
	if err != nil {
		return errors.New("error binding the UDP socket to the configured listen address: " + err.Error())
	}

	return nil
}

func createUDP(ipv6Enabled bool) (int, error) {
	if ipv6Enabled {
		return syscall.Socket(syscall.AF_INET6, syscall.SOCK_DGRAM, 0)
	}
	return syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
}
