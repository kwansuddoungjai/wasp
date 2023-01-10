package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type OffLedgerService struct {
	chainService    interfaces.ChainService
	networkProvider peering.NetworkProvider
	requestCache    *expiringcache.ExpiringCache
}

func NewOffLedgerService(chainService interfaces.ChainService, networkProvider peering.NetworkProvider, requestCacheTTL time.Duration) interfaces.OffLedgerService {
	return &OffLedgerService{
		chainService:    chainService,
		networkProvider: networkProvider,
		requestCache:    expiringcache.New(requestCacheTTL),
	}
}

func (c *OffLedgerService) ParseRequest(binaryRequest []byte) (isc.OffLedgerRequest, error) {
	request, err := isc.NewRequestFromMarshalUtil(marshalutil.New(binaryRequest))
	if err != nil {
		return nil, errors.New("error parsing request from payload")
	}

	req, ok := request.(isc.OffLedgerRequest)
	if !ok {
		return nil, errors.New("error parsing request: off-ledger request expected")
	}

	return req, nil
}

func (c *OffLedgerService) EnqueueOffLedgerRequest(chainID isc.ChainID, binaryRequest []byte) error {
	request, err := c.ParseRequest(binaryRequest)
	if err != nil {
		return err
	}

	reqID := request.ID()

	if c.requestCache.Get(reqID) != nil {
		return fmt.Errorf("request already processed")
	}

	// check req signature
	if err := request.VerifySignature(); err != nil {
		c.requestCache.Set(reqID, true)
		return fmt.Errorf("could not verify: %s", err.Error())
	}

	// check req is for the correct chain
	if !request.ChainID().Equals(chainID) {
		// do not add to cache, it can still be sent to the correct chain
		return fmt.Errorf("Request is for a different chain")
	}

	// check chain exists
	chain := c.chainService.GetChainByID(chainID)
	if chain == nil {
		return fmt.Errorf("Unknown chain: %s", chainID.String())
	}

	if err := ShouldBeProcessed(chain, request); err != nil {
		return err
	}

	chain.ReceiveOffLedgerRequest(request, c.networkProvider.Self().PubKey())

	return nil
}

// implemented this way so we can re-use the same state, and avoid the overhead of calling views
// TODO exported just to be used by V1 API until that gets deprecated at once.
func ShouldBeProcessed(ch chain.ChainCore, req isc.OffLedgerRequest) error {
	state, err := ch.GetStateReader().LatestState()
	if err != nil {
		return httperrors.ServerError("unable to get latest state")
	}

	// query blocklog contract
	blocklogPartition := subrealm.NewReadOnly(state, kv.Key(blocklog.Contract.Hname().Bytes()))
	receipt, err := blocklog.IsRequestProcessedInternal(blocklogPartition, req.ID())
	if err != nil {
		return httperrors.ServerError("unable to get request receipt from block state")
	}
	if receipt != nil {
		return httperrors.BadRequest("request already processed")
	}

	// query accounts contract
	accountsPartition := subrealm.NewReadOnly(state, kv.Key(accounts.Contract.Hname().Bytes()))
	// check user has on-chain balance
	if !accounts.AccountExists(accountsPartition, req.SenderAccount()) {
		return httperrors.BadRequest(fmt.Sprintf("No balance on account %s", req.SenderAccount().String()))
	}
	accountNonce := accounts.GetMaxAssumedNonce(accountsPartition, req.SenderAccount())
	if err := vmcontext.CheckNonce(req, accountNonce); err != nil {
		return httperrors.BadRequest(fmt.Sprintf("invalid nonce, %v", err))
	}
	return nil
}
