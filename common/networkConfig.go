// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"encoding/json"
	"errors"
	"net"
	"time"
)

// DefaultNetworkConfig to use when the NetworkConfig is not specified in the backend datastore.
var DefaultNetworkConfig *NetworkConfig

// NetworkConfig object to represent the current network setup.
type NetworkConfig struct {
	// The network range that represents the quantum network.
	Network string `json:"network"`

	// The reserved static ip address range which should be skipped for DHCP assignments.
	StaticRange string `json:"staticRange"`

	// The length of time to hold the assigned DHCP lease.
	LeaseTime time.Duration `json:"leaseTime"`

	// The base ip address of the quantum network.
	BaseIP net.IP `json:"-"`

	// The IPNet representation of the quantum network.
	IPNet *net.IPNet `json:"-"`

	// The IPNet representation of the reserved static ip address range.
	StaticNet *net.IPNet `json:"-"`
}

// ParseNetworkConfig from the data stored in the backend datastore.
func ParseNetworkConfig(data []byte) (*NetworkConfig, error) {
	var networkCfg NetworkConfig
	json.Unmarshal(data, &networkCfg)

	if networkCfg.LeaseTime == 0 {
		networkCfg.LeaseTime = 48 * time.Hour
	}

	baseIP, ipnet, err := net.ParseCIDR(networkCfg.Network)
	if err != nil {
		return nil, err
	}

	networkCfg.BaseIP = baseIP
	networkCfg.IPNet = ipnet

	if networkCfg.StaticRange == "" {
		return &networkCfg, nil
	}

	staticBase, staticNet, err := net.ParseCIDR(networkCfg.StaticRange)
	if err != nil {
		return nil, err
	} else if !ipnet.Contains(staticBase) {
		return nil, errors.New("network configuration has staticRange defined but the range does not exist in the configured network")
	}

	networkCfg.StaticNet = staticNet
	return &networkCfg, nil
}

// Bytes returns a byte slice representation of a NetworkConfig object, if there is an error while marshalling data a nil slice is returned.
func (networkCfg *NetworkConfig) Bytes() []byte {
	buf, _ := json.Marshal(networkCfg)
	return buf
}

// Bytes returns a string representation of a NetworkConfig object, if there is an error while marshalling data an empty string is returned.
func (networkCfg *NetworkConfig) String() string {
	return string(networkCfg.Bytes())
}

func init() {
	defaultLeaseTime, _ := time.ParseDuration("48h")
	DefaultNetworkConfig = &NetworkConfig{
		Network:     "10.99.0.0/16",
		StaticRange: "10.99.0.0/23",
		LeaseTime:   defaultLeaseTime,
	}

	baseIP, ipnet, _ := net.ParseCIDR(DefaultNetworkConfig.Network)
	DefaultNetworkConfig.BaseIP = baseIP
	DefaultNetworkConfig.IPNet = ipnet

	_, staticNet, _ := net.ParseCIDR(DefaultNetworkConfig.StaticRange)
	DefaultNetworkConfig.StaticNet = staticNet
}
