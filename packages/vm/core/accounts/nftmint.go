package accounts

import (
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

type mintedNFTRecord struct {
	positionInMintedList uint16
	outputIndex          uint16
	owner                isc.AgentID
	output               *iotago.NFTOutput
	nft                  *isc.NFT
}

func (rec *mintedNFTRecord) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rec.positionInMintedList = rr.ReadUint16()
	rec.outputIndex = rr.ReadUint16()
	rec.owner = isc.AgentIDFromReader(rr)
	rec.output = new(iotago.NFTOutput)
	rr.ReadSerialized(rec.output)
	rec.nft = new(isc.NFT)
	rr.Read(rec.nft)
	return rr.Err
}

func (rec *mintedNFTRecord) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteUint16(rec.positionInMintedList)
	ww.WriteUint16(rec.outputIndex)
	if rec.owner != nil {
		ww.Write(rec.owner)
	} else {
		ww.Write(&isc.NilAgentID{})
	}
	ww.WriteSerialized(rec.output)
	ww.Write(rec.nft)
	return ww.Err
}

func (rec *mintedNFTRecord) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

func mintedNFTRecordFromBytes(data []byte) *mintedNFTRecord {
	record, err := rwutil.ReadFromBytes(data, new(mintedNFTRecord))
	if err != nil {
		panic(err)
	}
	return record
}

func newlyMintedNFTsMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixNewlyMintedNFTs)
}

func mintIDMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, prefixInternalNFTIDMap)
}

func mintIDMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, prefixInternalNFTIDMap)
}

var (
	errMintNFTWithdraw = coreerrors.Register("can only withdraw on mint to a L1 address").Create()
	errInvalidAgentID  = coreerrors.Register("invalid agentID").Create()
)

func mintParams(ctx isc.Sandbox) (iotago.Address, isc.AgentID, []byte) {
	params := ctx.Params()

	immutableMetadata := params.MustGetBytes(ParamNFTImmutableData)
	targetAgentID := params.MustGetAgentID(ParamAgentID)
	withdrawOnMint := params.MustGetBool(ParamNFTWithdrawOnMint, false)

	chainAddress := ctx.ChainID().AsAddress()
	switch targetAgentID.Kind() {
	case isc.AgentIDKindContract, isc.AgentIDKindEthereumAddress:
		if withdrawOnMint {
			panic(errMintNFTWithdraw)
		}
		return chainAddress, targetAgentID, immutableMetadata
	case isc.AgentIDKindAddress:
		if withdrawOnMint {
			targetAddress := targetAgentID.(*isc.AddressAgentID).Address()
			return targetAddress, nil, immutableMetadata
		}
		return chainAddress, targetAgentID, immutableMetadata
	default:
		panic(errInvalidAgentID)
	}
}

func internalNFTID(blockIndex uint32, positionInMintedList uint16) []byte {
	ret := make([]byte, 6)
	copy(ret[0:], codec.EncodeUint32(blockIndex))
	copy(ret[4:], codec.EncodeUint16(positionInMintedList))
	return ret
}

// NFTs are always minted with the minimumSD and that must be provided via allowance
func mintNFT(ctx isc.Sandbox) dict.Dict {
	// TODO can we even do "NFT collections"?
	targetAddress, ownerAgentID, immutableMetadata := mintParams(ctx)

	positionInMintedList, nftOutput := ctx.Privileged().MintNFT(targetAddress, immutableMetadata)

	// debit the SD required for the NFT from the sender account
	ctx.TransferAllowedFunds(ctx.AccountID(), isc.NewAssetsBaseTokens(nftOutput.Amount))      // claim tokens from allowance
	DebitFromAccount(ctx.State(), ctx.AccountID(), isc.NewAssetsBaseTokens(nftOutput.Amount)) // debit from this SC account

	rec := mintedNFTRecord{
		positionInMintedList: positionInMintedList,
		outputIndex:          0,
		owner:                ownerAgentID,
		output:               nftOutput,
		nft: &isc.NFT{
			ID:       [32]byte{},
			Issuer:   ctx.ChainID().AsAddress(),
			Metadata: immutableMetadata,
			Owner:    ownerAgentID,
		},
	}
	// save the info required to credit the NFT on next block
	newlyMintedNFTsMap(ctx.State()).SetAt(codec.Encode(positionInMintedList), rec.Bytes())

	return dict.Dict{
		ParamInternalMintID: internalNFTID(ctx.StateAnchor().StateIndex+1, positionInMintedList),
	}
}

func viewNFTIDbyMintID(ctx isc.SandboxView) dict.Dict {
	internalMintID := ctx.Params().MustGetBytes(ParamInternalMintID)
	return dict.Dict{
		ParamNFTID: mintIDMapR(ctx.StateR()).GetAt(internalMintID),
	}
}

// ----  output management

func SaveMintedNFTOutput(state kv.KVStore, positionInMintedList, outputIndex uint16) {
	mintMap := newlyMintedNFTsMap(state)
	key := codec.Encode(positionInMintedList)
	recBytes := mintMap.GetAt(key)
	if recBytes == nil {
		return
	}
	rec := mintedNFTRecordFromBytes(recBytes)
	rec.outputIndex = outputIndex
	mintMap.SetAt(key, rec.Bytes())
}

func updateNewlyMintedNFTOutputIDs(state kv.KVStore, anchorTxID iotago.TransactionID, blockIndex uint32) {
	mintMap := newlyMintedNFTsMap(state)
	nftMap := NFTOutputMap(state)
	mintMap.Iterate(func(_, recBytes []byte) bool {
		mintedRec := mintedNFTRecordFromBytes(recBytes)
		// calculate the NFTID from the anchor txID	+ outputIndex
		outputID := iotago.OutputIDFromTransactionIDAndIndex(anchorTxID, mintedRec.outputIndex)
		nftID := iotago.NFTIDFromOutputID(outputID)

		if mintedRec.owner.Kind() != isc.AgentIDKindNil { // when owner is nil, means the NFT was minted directly to a L1 wallet
			mintedRec.output.NFTID = nftID
			mintedRec.nft.ID = nftID
			outputRec := NFTOutputRec{
				OutputID: outputID,
				Output:   mintedRec.output,
			}
			// save the updated data in the NFT map
			nftMap.SetAt(nftID[:], outputRec.Bytes())
			// credit the NFT to the target owner
			creditNFTToAccount(state, mintedRec.owner, mintedRec.nft)
		}
		// save the mapping of [internalID => NFTID]
		mintIDMap(state).SetAt(internalNFTID(blockIndex, mintedRec.positionInMintedList), nftID[:])
		return true
	})
	mintMap.Erase()
}
