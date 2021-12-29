// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// ChainInfo is an API structure which contains main properties of the chain in on place
type ChainInfo struct {
	ChainID         *iscp.ChainID
	ChainOwnerID    *iscp.AgentID
	Description     string
	GasFeePolicy    *gas.GasFeePolicy
	MaxBlobSize     uint32
	MaxEventSize    uint16
	MaxEventsPerReq uint16
}
