// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

/*
Package workers contains the structs and logic to handle routing, encrypting, and analyzing traffic over the quantum network. Each worker handles one direction of network traffic, either incoming traffic from remote nodes or outgoing traffic destined for remote nodes.
*/
package workers

import (
	"github.com/Supernomad/quantum/common"
)

func isTrusted(cfg *common.Config, mapping *common.Mapping) bool {
	if cfg.TrustedNetworks == nil {
		return false
	}

	for i := 0; i < len(cfg.TrustedNetworks); i++ {
		if cfg.TrustedNetworks[i].Contains(mapping.PrivateIP) {
			return true
		}
	}
	return false
}
