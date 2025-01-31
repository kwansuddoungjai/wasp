package isc

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/minio/blake2b-simd"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type offLedgerSignature struct {
	publicKey *cryptolib.PublicKey
	signature []byte
}

type OffLedgerRequestData struct {
	allowance  *Assets
	chainID    ChainID
	contract   Hname
	entryPoint Hname
	gasBudget  uint64
	nonce      uint64
	params     dict.Dict
	signature  offLedgerSignature
}

var (
	_ Request                  = new(OffLedgerRequestData)
	_ OffLedgerRequest         = new(OffLedgerRequestData)
	_ UnsignedOffLedgerRequest = new(OffLedgerRequestData)
	_ Calldata                 = new(OffLedgerRequestData)
	_ Features                 = new(OffLedgerRequestData)
)

type ImpersonatedOffLedgerRequestData struct {
	req     *OffLedgerRequestData
	address *iotago.Ed25519Address
}

func NewImpersonatedOffLedgerRequest(request OffLedgerRequest) ImpersonatedOffLedgerRequest {
	gasBudget, _ := request.GasBudget()

	return &ImpersonatedOffLedgerRequestData{
		req: &OffLedgerRequestData{
			allowance:  request.Allowance().Clone(),
			chainID:    request.ChainID(),
			contract:   request.CallTarget().Contract,
			entryPoint: request.CallTarget().EntryPoint,
			gasBudget:  gasBudget,
			nonce:      request.Nonce(),
			params:     request.Params().Clone(),
			signature: offLedgerSignature{
				signature: make([]byte, 0),
				publicKey: cryptolib.NewEmptyPublicKey(),
			},
		},
	}
}

func (r *ImpersonatedOffLedgerRequestData) WithSenderAddress(address *iotago.Ed25519Address) OffLedgerRequest {
	r.address = address
	return r
}

func (r *ImpersonatedOffLedgerRequestData) Allowance() *Assets {
	return r.req.Allowance()
}

func (r *ImpersonatedOffLedgerRequestData) Assets() *Assets {
	return r.req.Assets()
}

func (r *ImpersonatedOffLedgerRequestData) CallTarget() CallTarget {
	return r.req.CallTarget()
}

func (r *ImpersonatedOffLedgerRequestData) GasBudget() (gas uint64, isEVM bool) {
	return r.req.GasBudget()
}

func (r *ImpersonatedOffLedgerRequestData) ID() RequestID {
	return r.req.ID()
}

func (r *ImpersonatedOffLedgerRequestData) NFT() *NFT {
	return r.req.NFT()
}

func (r *ImpersonatedOffLedgerRequestData) Params() dict.Dict {
	return r.req.Params()
}

func (r *ImpersonatedOffLedgerRequestData) SenderAccount() AgentID {
	return NewAgentID(r.address)
}

func (r *ImpersonatedOffLedgerRequestData) TargetAddress() iotago.Address {
	return r.req.TargetAddress()
}

func (r *ImpersonatedOffLedgerRequestData) TxValue() *big.Int {
	return r.req.TxValue()
}

func (r *ImpersonatedOffLedgerRequestData) EVMCallData() *EVMCallData {
	return r.req.EVMCallData()
}

func (r *ImpersonatedOffLedgerRequestData) Bytes() []byte {
	return r.req.Bytes()
}

func (r *ImpersonatedOffLedgerRequestData) IsOffLedger() bool {
	return r.req.IsOffLedger()
}

func (r *ImpersonatedOffLedgerRequestData) String() string {
	return r.req.String()
}

func (r *ImpersonatedOffLedgerRequestData) Read(reader io.Reader) error {
	return r.req.Read(reader)
}

func (r *ImpersonatedOffLedgerRequestData) Write(w io.Writer) error {
	return r.req.Write(w)
}

func (r *ImpersonatedOffLedgerRequestData) ChainID() ChainID {
	return r.req.ChainID()
}

func (r *ImpersonatedOffLedgerRequestData) Nonce() uint64 {
	return r.req.Nonce()
}

func (r *ImpersonatedOffLedgerRequestData) VerifySignature() error {
	return r.req.VerifySignature()
}

func (r *ImpersonatedOffLedgerRequestData) EVMTransaction() *types.Transaction {
	return r.req.EVMTransaction()
}

func NewOffLedgerRequest(
	chainID ChainID,
	contract, entryPoint Hname,
	params dict.Dict,
	nonce uint64,
	gasBudget uint64,
) UnsignedOffLedgerRequest {
	return &OffLedgerRequestData{
		chainID:    chainID,
		contract:   contract,
		entryPoint: entryPoint,
		params:     params,
		nonce:      nonce,
		allowance:  NewEmptyAssets(),
		gasBudget:  gasBudget,
	}
}

func (req *OffLedgerRequestData) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	req.readEssence(rr)
	req.signature.publicKey = cryptolib.NewEmptyPublicKey()
	rr.Read(req.signature.publicKey)
	req.signature.signature = rr.ReadBytes()
	return rr.Err
}

func (req *OffLedgerRequestData) TxValue() *big.Int {
	return big.NewInt(0)
}

func (req *OffLedgerRequestData) EVMCallData() *EVMCallData {
	return nil
}

func (req *OffLedgerRequestData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	req.writeEssence(ww)
	ww.Write(req.signature.publicKey)
	ww.WriteBytes(req.signature.signature)
	return ww.Err
}

func (req *OffLedgerRequestData) readEssence(rr *rwutil.Reader) {
	rr.ReadKindAndVerify(rwutil.Kind(requestKindOffLedgerISC))
	rr.Read(&req.chainID)
	rr.Read(&req.contract)
	rr.Read(&req.entryPoint)
	req.params = dict.New()
	rr.Read(&req.params)
	req.nonce = rr.ReadAmount64()
	req.gasBudget = rr.ReadGas64()
	req.allowance = NewEmptyAssets()
	rr.Read(req.allowance)
}

func (req *OffLedgerRequestData) writeEssence(ww *rwutil.Writer) {
	ww.WriteKind(rwutil.Kind(requestKindOffLedgerISC))
	ww.Write(&req.chainID)
	ww.Write(&req.contract)
	ww.Write(&req.entryPoint)
	ww.Write(&req.params)
	ww.WriteAmount64(req.nonce)
	ww.WriteGas64(req.gasBudget)
	ww.Write(req.allowance)
}

// Allowance from the sender's account to the target smart contract. Nil mean no Allowance
func (req *OffLedgerRequestData) Allowance() *Assets {
	return req.allowance
}

// Assets is attached assets to the UTXO. Nil for off-ledger
func (req *OffLedgerRequestData) Assets() *Assets {
	return nil
}

func (req *OffLedgerRequestData) Bytes() []byte {
	return rwutil.WriteToBytes(req)
}

func (req *OffLedgerRequestData) CallTarget() CallTarget {
	return CallTarget{
		Contract:   req.contract,
		EntryPoint: req.entryPoint,
	}
}

func (req *OffLedgerRequestData) ChainID() ChainID {
	return req.chainID
}

func (req *OffLedgerRequestData) EssenceBytes() []byte {
	ww := rwutil.NewBytesWriter()
	req.writeEssence(ww)
	return ww.Bytes()
}

func (req *OffLedgerRequestData) messageToSign() []byte {
	ret := blake2b.Sum256(req.EssenceBytes())
	return ret[:]
}

func (req *OffLedgerRequestData) Expiry() (time.Time, iotago.Address) {
	return time.Time{}, nil
}

func (req *OffLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	return req.gasBudget, false
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (req *OffLedgerRequestData) ID() (requestID RequestID) {
	return NewRequestID(iotago.TransactionID(hashing.HashData(req.Bytes())), 0)
}

func (req *OffLedgerRequestData) IsOffLedger() bool {
	return true
}

func (req *OffLedgerRequestData) NFT() *NFT {
	return nil
}

// Nonce incremental nonce used for replay protection
func (req *OffLedgerRequestData) Nonce() uint64 {
	return req.nonce
}

func (req *OffLedgerRequestData) Params() dict.Dict {
	return req.params
}

func (req *OffLedgerRequestData) ReturnAmount() (uint64, bool) {
	return 0, false
}

func (req *OffLedgerRequestData) SenderAccount() AgentID {
	return NewAgentID(req.signature.publicKey.AsEd25519Address())
}

// Sign signs the essence
func (req *OffLedgerRequestData) Sign(key cryptolib.VariantKeyPair) OffLedgerRequest {
	req.signature = offLedgerSignature{
		publicKey: key.GetPublicKey(),
		signature: key.SignBytes(req.messageToSign()),
	}
	return req
}

func (req *OffLedgerRequestData) String() string {
	return fmt.Sprintf("offLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, nonce: %d }",
		req.ID().String(),
		req.SenderAccount().String(),
		req.contract.String(),
		req.entryPoint.String(),
		req.Params().String(),
		req.nonce,
	)
}

func (req *OffLedgerRequestData) TargetAddress() iotago.Address {
	return req.chainID.AsAddress()
}

func (req *OffLedgerRequestData) TimeLock() time.Time {
	return time.Time{}
}

func (req *OffLedgerRequestData) Timestamp() time.Time {
	// no request TX, return zero time
	return time.Time{}
}

// VerifySignature verifies essence signature
func (req *OffLedgerRequestData) VerifySignature() error {
	if !req.signature.publicKey.Verify(req.messageToSign(), req.signature.signature) {
		return errors.New("invalid signature")
	}
	return nil
}

func (req *OffLedgerRequestData) WithAllowance(allowance *Assets) UnsignedOffLedgerRequest {
	req.allowance = allowance.Clone()
	return req
}

func (req *OffLedgerRequestData) WithGasBudget(gasBudget uint64) UnsignedOffLedgerRequest {
	req.gasBudget = gasBudget
	return req
}

func (req *OffLedgerRequestData) WithNonce(nonce uint64) UnsignedOffLedgerRequest {
	req.nonce = nonce
	return req
}

// WithSender can be used to estimate gas, without a signature
func (req *OffLedgerRequestData) WithSender(sender *cryptolib.PublicKey) UnsignedOffLedgerRequest {
	req.signature = offLedgerSignature{
		publicKey: sender,
		signature: []byte{},
	}
	return req
}

func (*OffLedgerRequestData) EVMTransaction() *types.Transaction {
	return nil
}
