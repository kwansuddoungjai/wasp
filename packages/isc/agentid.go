// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"errors"
	"io"
	"strings"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type AgentIDKind rwutil.Kind

const (
	AgentIDKindNil AgentIDKind = iota
	AgentIDKindAddress
	AgentIDKindContract
	AgentIDKindEthereumAddress
)

// AgentID represents any entity that can hold assets on L2 and/or call contracts.
type AgentID interface {
	Bytes() []byte
	Equals(other AgentID) bool
	Kind() AgentIDKind
	Read(r io.Reader) error
	String() string
	Write(w io.Writer) error
}

// AgentIDWithL1Address is an AgentID backed by an L1 address (either AddressAgentID or ContractAgentID).
type AgentIDWithL1Address interface {
	AgentID
	Address() iotago.Address
}

// AddressFromAgentID returns the L1 address of the AgentID, if applicable.
func AddressFromAgentID(a AgentID) (iotago.Address, bool) {
	wa, ok := a.(AgentIDWithL1Address)
	if !ok {
		return nil, false
	}
	return wa.Address(), true
}

// HnameFromAgentID returns the hname of the AgentID, or HnameNil if not applicable.
func HnameFromAgentID(a AgentID) Hname {
	if ca, ok := a.(*ContractAgentID); ok {
		return ca.Hname()
	}
	return HnameNil
}

// NewAgentID creates an AddressAgentID if the address is not an AliasAddress;
// otherwise a ContractAgentID with hname = HnameNil.
func NewAgentID(addr iotago.Address) AgentID {
	if addr.Type() == iotago.AddressAlias {
		chainID := ChainIDFromAddress(addr.(*iotago.AliasAddress))
		return NewContractAgentID(chainID, 0)
	}
	return &AddressAgentID{a: addr}
}

func AgentIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (AgentID, error) {
	rr := rwutil.NewMuReader(mu)
	return agentIDFromReader(rr), rr.Err
}

func AgentIDFromBytes(data []byte) (AgentID, error) {
	rr := rwutil.NewBytesReader(data)
	return agentIDFromReader(rr), rr.Err
}

func agentIDFromReader(rr *rwutil.Reader) (ret AgentID) {
	kind := rr.ReadKind()
	switch AgentIDKind(kind) {
	case AgentIDKindNil:
		ret = new(NilAgentID)
	case AgentIDKindAddress:
		ret = new(AddressAgentID)
	case AgentIDKindContract:
		ret = new(ContractAgentID)
	case AgentIDKindEthereumAddress:
		ret = new(EthereumAddressAgentID)
	default:
		if rr.Err == nil {
			rr.Err = errors.New("invalid AgentID kind")
			return nil
		}
	}
	rr.PushBack().WriteKind(kind)
	rr.Read(ret)
	return ret
}

// NewAgentIDFromString parses the human-readable string representation
func NewAgentIDFromString(s string) (AgentID, error) {
	if s == nilAgentIDString {
		return &NilAgentID{}, nil
	}
	var hnamePart, addrPart string
	{
		parts := strings.Split(s, "@")
		switch len(parts) {
		case 1:
			addrPart = parts[0]
		case 2:
			addrPart = parts[1]
			hnamePart = parts[0]
		default:
			return nil, errors.New("invalid AgentID format")
		}
	}

	if hnamePart != "" {
		return contractAgentIDFromString(hnamePart, addrPart)
	}
	if strings.HasPrefix(addrPart, string(parameters.L1().Protocol.Bech32HRP)) {
		return addressAgentIDFromString(s)
	}
	if strings.HasPrefix(addrPart, "0x") {
		return ethAgentIDFromString(s)
	}
	return nil, errors.New("invalid AgentID string")
}

// NewRandomAgentID creates random AgentID
func NewRandomAgentID() AgentID {
	return NewContractAgentID(RandomChainID(), Hn("testName"))
}
