// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

package coreblocklog

import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"

type ImmutableControlAddressesResults struct {
	id int32
}

func (s ImmutableControlAddressesResults) BlockIndex() wasmlib.ScImmutableInt32 {
	return wasmlib.NewScImmutableInt32(s.id, wasmlib.KeyID(ResultBlockIndex))
}

func (s ImmutableControlAddressesResults) GoverningAddress() wasmlib.ScImmutableAddress {
	return wasmlib.NewScImmutableAddress(s.id, wasmlib.KeyID(ResultGoverningAddress))
}

func (s ImmutableControlAddressesResults) StateControllerAddress() wasmlib.ScImmutableAddress {
	return wasmlib.NewScImmutableAddress(s.id, wasmlib.KeyID(ResultStateControllerAddress))
}

type MutableControlAddressesResults struct {
	id int32
}

func (s MutableControlAddressesResults) BlockIndex() wasmlib.ScMutableInt32 {
	return wasmlib.NewScMutableInt32(s.id, wasmlib.KeyID(ResultBlockIndex))
}

func (s MutableControlAddressesResults) GoverningAddress() wasmlib.ScMutableAddress {
	return wasmlib.NewScMutableAddress(s.id, wasmlib.KeyID(ResultGoverningAddress))
}

func (s MutableControlAddressesResults) StateControllerAddress() wasmlib.ScMutableAddress {
	return wasmlib.NewScMutableAddress(s.id, wasmlib.KeyID(ResultStateControllerAddress))
}

type ImmutableGetBlockInfoResults struct {
	id int32
}

func (s ImmutableGetBlockInfoResults) BlockInfo() wasmlib.ScImmutableBytes {
	return wasmlib.NewScImmutableBytes(s.id, wasmlib.KeyID(ResultBlockInfo))
}

type MutableGetBlockInfoResults struct {
	id int32
}

func (s MutableGetBlockInfoResults) BlockInfo() wasmlib.ScMutableBytes {
	return wasmlib.NewScMutableBytes(s.id, wasmlib.KeyID(ResultBlockInfo))
}

type ArrayOfImmutableBytes struct {
	objID int32
}

func (a ArrayOfImmutableBytes) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfImmutableBytes) GetBytes(index int32) wasmlib.ScImmutableBytes {
	return wasmlib.NewScImmutableBytes(a.objID, wasmlib.Key32(index))
}

type ImmutableGetEventsForBlockResults struct {
	id int32
}

func (s ImmutableGetEventsForBlockResults) Event() ArrayOfImmutableBytes {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultEvent), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_BYTES)
	return ArrayOfImmutableBytes{objID: arrID}
}

type ArrayOfMutableBytes struct {
	objID int32
}

func (a ArrayOfMutableBytes) Clear() {
	wasmlib.Clear(a.objID)
}

func (a ArrayOfMutableBytes) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfMutableBytes) GetBytes(index int32) wasmlib.ScMutableBytes {
	return wasmlib.NewScMutableBytes(a.objID, wasmlib.Key32(index))
}

type MutableGetEventsForBlockResults struct {
	id int32
}

func (s MutableGetEventsForBlockResults) Event() ArrayOfMutableBytes {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultEvent), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_BYTES)
	return ArrayOfMutableBytes{objID: arrID}
}

type ImmutableGetEventsForContractResults struct {
	id int32
}

func (s ImmutableGetEventsForContractResults) Event() ArrayOfImmutableBytes {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultEvent), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_BYTES)
	return ArrayOfImmutableBytes{objID: arrID}
}

type MutableGetEventsForContractResults struct {
	id int32
}

func (s MutableGetEventsForContractResults) Event() ArrayOfMutableBytes {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultEvent), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_BYTES)
	return ArrayOfMutableBytes{objID: arrID}
}

type ImmutableGetEventsForRequestResults struct {
	id int32
}

func (s ImmutableGetEventsForRequestResults) Event() ArrayOfImmutableBytes {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultEvent), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_BYTES)
	return ArrayOfImmutableBytes{objID: arrID}
}

type MutableGetEventsForRequestResults struct {
	id int32
}

func (s MutableGetEventsForRequestResults) Event() ArrayOfMutableBytes {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultEvent), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_BYTES)
	return ArrayOfMutableBytes{objID: arrID}
}

type ImmutableGetLatestBlockInfoResults struct {
	id int32
}

func (s ImmutableGetLatestBlockInfoResults) BlockIndex() wasmlib.ScImmutableInt32 {
	return wasmlib.NewScImmutableInt32(s.id, wasmlib.KeyID(ResultBlockIndex))
}

func (s ImmutableGetLatestBlockInfoResults) BlockInfo() wasmlib.ScImmutableBytes {
	return wasmlib.NewScImmutableBytes(s.id, wasmlib.KeyID(ResultBlockInfo))
}

type MutableGetLatestBlockInfoResults struct {
	id int32
}

func (s MutableGetLatestBlockInfoResults) BlockIndex() wasmlib.ScMutableInt32 {
	return wasmlib.NewScMutableInt32(s.id, wasmlib.KeyID(ResultBlockIndex))
}

func (s MutableGetLatestBlockInfoResults) BlockInfo() wasmlib.ScMutableBytes {
	return wasmlib.NewScMutableBytes(s.id, wasmlib.KeyID(ResultBlockInfo))
}

type ArrayOfImmutableRequestID struct {
	objID int32
}

func (a ArrayOfImmutableRequestID) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfImmutableRequestID) GetRequestID(index int32) wasmlib.ScImmutableRequestID {
	return wasmlib.NewScImmutableRequestID(a.objID, wasmlib.Key32(index))
}

type ImmutableGetRequestIDsForBlockResults struct {
	id int32
}

func (s ImmutableGetRequestIDsForBlockResults) RequestID() ArrayOfImmutableRequestID {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultRequestID), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_REQUEST_ID)
	return ArrayOfImmutableRequestID{objID: arrID}
}

type ArrayOfMutableRequestID struct {
	objID int32
}

func (a ArrayOfMutableRequestID) Clear() {
	wasmlib.Clear(a.objID)
}

func (a ArrayOfMutableRequestID) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfMutableRequestID) GetRequestID(index int32) wasmlib.ScMutableRequestID {
	return wasmlib.NewScMutableRequestID(a.objID, wasmlib.Key32(index))
}

type MutableGetRequestIDsForBlockResults struct {
	id int32
}

func (s MutableGetRequestIDsForBlockResults) RequestID() ArrayOfMutableRequestID {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultRequestID), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_REQUEST_ID)
	return ArrayOfMutableRequestID{objID: arrID}
}

type ImmutableGetRequestReceiptResults struct {
	id int32
}

func (s ImmutableGetRequestReceiptResults) BlockIndex() wasmlib.ScImmutableInt32 {
	return wasmlib.NewScImmutableInt32(s.id, wasmlib.KeyID(ResultBlockIndex))
}

func (s ImmutableGetRequestReceiptResults) RequestIndex() wasmlib.ScImmutableInt16 {
	return wasmlib.NewScImmutableInt16(s.id, wasmlib.KeyID(ResultRequestIndex))
}

func (s ImmutableGetRequestReceiptResults) RequestRecord() wasmlib.ScImmutableBytes {
	return wasmlib.NewScImmutableBytes(s.id, wasmlib.KeyID(ResultRequestRecord))
}

type MutableGetRequestReceiptResults struct {
	id int32
}

func (s MutableGetRequestReceiptResults) BlockIndex() wasmlib.ScMutableInt32 {
	return wasmlib.NewScMutableInt32(s.id, wasmlib.KeyID(ResultBlockIndex))
}

func (s MutableGetRequestReceiptResults) RequestIndex() wasmlib.ScMutableInt16 {
	return wasmlib.NewScMutableInt16(s.id, wasmlib.KeyID(ResultRequestIndex))
}

func (s MutableGetRequestReceiptResults) RequestRecord() wasmlib.ScMutableBytes {
	return wasmlib.NewScMutableBytes(s.id, wasmlib.KeyID(ResultRequestRecord))
}

type ImmutableGetRequestReceiptsForBlockResults struct {
	id int32
}

func (s ImmutableGetRequestReceiptsForBlockResults) RequestRecord() ArrayOfImmutableBytes {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultRequestRecord), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_BYTES)
	return ArrayOfImmutableBytes{objID: arrID}
}

type MutableGetRequestReceiptsForBlockResults struct {
	id int32
}

func (s MutableGetRequestReceiptsForBlockResults) RequestRecord() ArrayOfMutableBytes {
	arrID := wasmlib.GetObjectID(s.id, wasmlib.KeyID(ResultRequestRecord), wasmlib.TYPE_ARRAY16|wasmlib.TYPE_BYTES)
	return ArrayOfMutableBytes{objID: arrID}
}

type ImmutableIsRequestProcessedResults struct {
	id int32
}

func (s ImmutableIsRequestProcessedResults) RequestProcessed() wasmlib.ScImmutableString {
	return wasmlib.NewScImmutableString(s.id, wasmlib.KeyID(ResultRequestProcessed))
}

type MutableIsRequestProcessedResults struct {
	id int32
}

func (s MutableIsRequestProcessedResults) RequestProcessed() wasmlib.ScMutableString {
	return wasmlib.NewScMutableString(s.id, wasmlib.KeyID(ResultRequestProcessed))
}
