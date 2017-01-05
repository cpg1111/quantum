// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package workers

import (
	"crypto/rand"
	"encoding/binary"
	"net/rpc"
	"path"

	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/device"
	"github.com/Supernomad/quantum/socket"
)

// Outgoing packet struct for handleing packets coming in off of a Device struct which are destined for a Socket struct.
type Outgoing struct {
	log  *common.Logger
	cfg  *common.Config
	cli  *rpc.Client
	dev  device.Device
	sock socket.Socket
	stop bool
}

func (outgoing *Outgoing) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.Packet[16:20])

	var mapping common.Mapping
	err := outgoing.cli.Call("DatastoreServer.Mapping", dip, &mapping)
	if err != nil {
		return nil, nil, false
	}

	outgoing.log.Debug.Println("[OUTGOING]", "Mapping retrieved: ", mapping)

	if outgoing.cfg.IsIPv6Enabled && mapping.IPv6 != nil {
		payload.Sockaddr = mapping.SockaddrInet6
	} else if outgoing.cfg.IsIPv4Enabled && mapping.IPv4 != nil {
		payload.Sockaddr = mapping.SockaddrInet4
	} else {
		return nil, nil, false
	}
	copy(payload.IPAddress, outgoing.cfg.PrivateIP.To4())
	return payload, &mapping, true
}

func (outgoing *Outgoing) seal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := rand.Read(payload.Nonce)
	if err != nil {
		return nil, false
	}

	mapping.Cipher.Seal(payload.Packet[:0], payload.Nonce, payload.Packet, payload.IPAddress)
	return payload, true
}

func (outgoing *Outgoing) stats(dropped bool, queue int, payload *common.Payload, mapping *common.Mapping) {
	aggStat := &common.Stat{
		Queue:     queue,
		Direction: common.OutgoingStat,
		Dropped:   dropped,
	}

	if payload != nil {
		aggStat.Bytes += uint64(payload.Length)
	}

	if mapping != nil {
		aggStat.PrivateIP = mapping.PrivateIP.String()
	}

	var reply int
	outgoing.cli.Call("AggServer.Sink", aggStat, &reply)
}

func (outgoing *Outgoing) pipeline(buf []byte, queue int) bool {
	payload, ok := outgoing.dev.Read(buf, queue)
	if !ok {
		outgoing.stats(true, queue, payload, nil)
		return ok
	}
	payload, mapping, ok := outgoing.resolve(payload)
	if !ok {
		outgoing.stats(true, queue, payload, mapping)
		return ok
	}
	payload, ok = outgoing.seal(payload, mapping)
	if !ok {
		outgoing.stats(true, queue, payload, mapping)
		return ok
	}
	ok = outgoing.sock.Write(payload, queue)
	if !ok {
		outgoing.stats(true, queue, payload, mapping)
		return ok
	}
	outgoing.stats(false, queue, payload, mapping)
	return true
}

// Start handling packets.
func (outgoing *Outgoing) Start(queue int) {
	go func() {
		outgoing.log.Debug.Println("[OUTGOING]", "Started main outgoing thread.")
		buf := make([]byte, common.MaxPacketLength)
		for !outgoing.stop {
			outgoing.pipeline(buf, queue)
		}
	}()
}

// Stop handling packets and shutdown.
func (outgoing *Outgoing) Stop() {
	outgoing.stop = true
}

// NewOutgoing generates an Outgoing worker which once started will handle packets coming from the local node destined for remote nodes in the quantum network.
func NewOutgoing(log *common.Logger, cfg *common.Config, dev device.Device, sock socket.Socket) *Outgoing {
	client, err := rpc.DialHTTP("unix", path.Join(cfg.DataDir, "quantum.sock"))
	if err != nil {
		log.Error.Println("error reaching master process via rpc: " + err.Error())
		return nil
	}

	return &Outgoing{
		log:  log,
		cfg:  cfg,
		cli:  client,
		dev:  dev,
		sock: sock,
		stop: false,
	}
}
