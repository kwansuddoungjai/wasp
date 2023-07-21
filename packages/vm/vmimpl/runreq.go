package vmimpl

import (
	"fmt"
	"math"
	"runtime/debug"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

// runRequest processes a single isc.Request in the batch
func (vmctx *vmContext) runRequest(req isc.Request, requestIndex uint16, maintenanceMode bool) (
	res *vm.RequestResult,
	unprocessableToRetry []isc.OnLedgerRequest,
	err error,
) {
	reqctx := &requestContext{
		vm:               vmctx,
		req:              req,
		requestIndex:     requestIndex,
		entropy:          hashing.HashData(append(codec.EncodeUint16(requestIndex), vmctx.task.Entropy[:]...)),
		uncommittedState: buffered.NewBufferedKVStore(vmctx.stateDraft),
	}

	if vmctx.task.EnableGasBurnLogging {
		reqctx.gas.burnLog = gas.NewGasBurnLog()
	}

	initialGasBurnedTotal := vmctx.blockGas.burned
	initialGasFeeChargedTotal := vmctx.blockGas.feeCharged

	reqctx.uncommittedState.Set(
		kv.Key(coreutil.StatePrefixTimestamp),
		codec.EncodeTime(vmctx.stateDraft.Timestamp().Add(1*time.Nanosecond)),
	)

	if err = reqctx.earlyCheckReasonToSkip(maintenanceMode); err != nil {
		return nil, nil, err
	}
	vmctx.loadChainConfig()

	// at this point state update is empty
	// so far there were no panics except optimistic reader
	txsnapshot := vmctx.createTxBuilderSnapshot()

	var result *vm.RequestResult
	err = reqctx.catchRequestPanic(
		func() {
			// transfer all attached assets to the sender's account
			reqctx.creditAssetsToChain()
			// load gas and fee policy, calculate and set gas budget
			reqctx.prepareGasBudget()
			// run the contract program
			receipt, callRet := reqctx.callTheContract()
			vmctx.mustCheckTransactionSize()
			result = &vm.RequestResult{
				Request: req,
				Receipt: receipt,
				Return:  callRet,
			}
		},
	)
	if err != nil {
		// protocol exception triggered. Skipping the request. Rollback
		vmctx.restoreTxBuilderSnapshot(txsnapshot)
		vmctx.blockGas.burned = initialGasBurnedTotal
		vmctx.blockGas.feeCharged = initialGasFeeChargedTotal

		return nil, nil, err
	}

	reqctx.uncommittedState.Mutations().ApplyTo(vmctx.stateDraft)
	return result, reqctx.unprocessableToRetry, nil
}

func (vmctx *vmContext) payoutAgentID() isc.AgentID {
	var payoutAgentID isc.AgentID
	withContractState(vmctx.stateDraft, governance.Contract, func(s kv.KVStore) {
		payoutAgentID = governance.MustGetPayoutAgentID(s)
	})
	return payoutAgentID
}

// creditAssetsToChain credits L1 accounts with attached assets and accrues all of them to the sender's account on-chain
func (reqctx *requestContext) creditAssetsToChain() {
	req := reqctx.req
	if req.IsOffLedger() {
		// off ledger request does not bring any deposit
		return
	}
	// Consume the output. Adjustment in L2 is needed because of storage deposit in the internal UTXOs
	storageDepositNeeded := reqctx.vm.txbuilder.Consume(req.(isc.OnLedgerRequest))

	// if sender is specified, all assets goes to sender's sender
	// Otherwise it all goes to the common sender and panics is logged in the SC call
	sender := req.SenderAccount()
	if sender == nil {
		if req.IsOffLedger() {
			panic("nil sender on offledger requests should never happen")
		}
		// onleger request with no sender, send all assets to the payoutAddress
		payoutAgentID := reqctx.vm.payoutAgentID()
		creditNFTToAccount(reqctx.uncommittedState, payoutAgentID, req.NFT())
		creditToAccount(reqctx.uncommittedState, payoutAgentID, req.Assets())

		// debit any SD required for accounting UTXOs
		if storageDepositNeeded > 0 {
			debitFromAccount(reqctx.uncommittedState, payoutAgentID, isc.NewAssetsBaseTokens(storageDepositNeeded))
		}
		return
	}

	senderBaseTokens := req.Assets().BaseTokens + reqctx.GetBaseTokensBalance(sender)

	if senderBaseTokens < storageDepositNeeded {
		// user doesn't have enough funds to pay for the SD needs of this request
		panic(vmexceptions.ErrNotEnoughFundsForSD)
	}

	creditToAccount(reqctx.uncommittedState, sender, req.Assets())
	creditNFTToAccount(reqctx.uncommittedState, sender, req.NFT())
	if storageDepositNeeded > 0 {
		reqctx.sdCharged = storageDepositNeeded
		debitFromAccount(reqctx.uncommittedState, sender, isc.NewAssetsBaseTokens(storageDepositNeeded))
	}
}

func (reqctx *requestContext) catchRequestPanic(f func()) error {
	err := panicutil.CatchPanic(f)
	if err == nil {
		return nil
	}
	// catches protocol exception error which is not the request or contract fault
	// If it occurs, the request is just skipped
	if vmexceptions.IsSkipRequestException(err) {
		return err
	}
	// panic again with more information about the error
	panic(fmt.Errorf(
		"panic when running request #%d ID:%s, requestbytes:%s err:%w",
		reqctx.requestIndex,
		reqctx.req.ID(),
		iotago.EncodeHex(reqctx.req.Bytes()),
		err,
	))
}

// checkAllowance ensure there are enough funds to cover the specified allowance
// panics if not enough funds
func (reqctx *requestContext) checkAllowance() {
	if !reqctx.HasEnoughForAllowance(reqctx.req.SenderAccount(), reqctx.req.Allowance()) {
		panic(vm.ErrNotEnoughFundsForAllowance)
	}
}

func (reqctx *requestContext) shouldChargeGasFee() bool {
	if reqctx.req.SenderAccount() == nil {
		return false
	}
	if reqctx.req.SenderAccount().Equals(reqctx.vm.ChainOwnerID()) && reqctx.req.CallTarget().Contract == governance.Contract.Hname() {
		return false
	}
	return true
}

func (reqctx *requestContext) prepareGasBudget() {
	if !reqctx.shouldChargeGasFee() {
		return
	}
	reqctx.gasSetBudget(reqctx.calculateAffordableGasBudget())
}

// callTheContract runs the contract. It catches and processes all panics except the one which cancel the whole block
func (reqctx *requestContext) callTheContract() (receipt *blocklog.RequestReceipt, callRet dict.Dict) {
	// TODO: do not mutate vmContext's txbuilder
	txSnapshot := reqctx.vm.createTxBuilderSnapshot()
	stateSnapshot := reqctx.uncommittedState.Clone()

	rollback := func() {
		reqctx.vm.restoreTxBuilderSnapshot(txSnapshot)
		reqctx.uncommittedState = stateSnapshot
	}

	var callErr *isc.VMError
	func() {
		defer func() {
			panicErr := checkVMPluginPanic(recover())
			if panicErr == nil {
				return
			}
			callErr = panicErr
			reqctx.Debugf("recovered panic from contract call: %v", panicErr)
			if reqctx.vm.task.WillProduceBlock() {
				reqctx.Debugf(string(debug.Stack()))
			}
		}()
		// ensure there are enough funds to cover the specified allowance
		reqctx.checkAllowance()

		reqctx.GasBurnEnable(true)
		callRet = reqctx.callFromRequest()
		// ensure at least the minimum amount of gas is charged
		reqctx.GasBurn(gas.BurnCodeMinimumGasPerRequest1P, reqctx.GasBurned())
	}()
	reqctx.GasBurnEnable(false)

	// execution over, save receipt, update nonces, etc
	// if anything goes wrong here, state must be rolled back and the request must be skipped
	func() {
		defer func() {
			if r := recover(); r != nil {
				rollback()
				callErrStr := ""
				if callErr != nil {
					callErrStr = callErr.Error()
				}
				reqctx.vm.task.Log.Errorf("panic after request execution (reqid: %s, executionErr: %s): %v", reqctx.req.ID(), callErrStr, r)
				reqctx.vm.task.Log.Debug(string(debug.Stack()))
				panic(vmexceptions.ErrPostExecutionPanic)
			}
		}()
		if callErr != nil {
			// panic happened during VM plugin call. Restore the state
			rollback()
		}
		// charge gas fee no matter what
		reqctx.chargeGasFee()

		// write receipt no matter what
		receipt = reqctx.writeReceiptToBlockLog(callErr)

		if reqctx.req.IsOffLedger() {
			reqctx.updateOffLedgerRequestNonce()
		}
	}()

	return receipt, callRet
}

func checkVMPluginPanic(r interface{}) *isc.VMError {
	if r == nil {
		return nil
	}
	// re-panic-ing if error it not user nor VM plugin fault.
	if vmexceptions.IsSkipRequestException(r) {
		panic(r)
	}
	// Otherwise, the panic is wrapped into the returned error, including gas-related panic
	switch err := r.(type) {
	case *isc.VMError:
		return r.(*isc.VMError)
	case isc.VMError:
		e := r.(isc.VMError)
		return &e
	case *kv.DBError:
		panic(err)
	case string:
		return coreerrors.ErrUntypedError.Create(err)
	case error:
		return coreerrors.ErrUntypedError.Create(err.Error())
	}
	return nil
}

// callFromRequest is the call itself. Assumes sc exists
func (reqctx *requestContext) callFromRequest() dict.Dict {
	req := reqctx.req
	reqctx.Debugf("callFromRequest: %s", req.ID().String())

	if req.SenderAccount() == nil {
		// if sender unknown, follow panic path
		panic(vm.ErrSenderUnknown)
	}

	contract := req.CallTarget().Contract
	entryPoint := req.CallTarget().EntryPoint

	return reqctx.callProgram(
		contract,
		entryPoint,
		req.Params(),
		req.Allowance(),
		req.SenderAccount(),
	)
}

func (reqctx *requestContext) getGasBudget() uint64 {
	gasBudget, isEVM := reqctx.req.GasBudget()
	if !isEVM || gasBudget == 0 {
		return gasBudget
	}

	var gasRatio util.Ratio32
	reqctx.callCore(governance.Contract, func(s kv.KVStore) {
		gasRatio = governance.MustGetGasFeePolicy(s).EVMGasRatio
	})
	return gas.EVMGasToISC(gasBudget, &gasRatio)
}

// calculateAffordableGasBudget checks the account of the sender and calculates affordable gas budget
// Affordable gas budget is calculated from gas budget provided in the request by the user and taking into account
// how many tokens the sender has in its account and how many are allowed for the target.
// Safe arithmetics is used
func (reqctx *requestContext) calculateAffordableGasBudget() (budget, maxTokensToSpendForGasFee uint64) {
	gasBudget := reqctx.getGasBudget()

	if reqctx.vm.task.EstimateGasMode && gasBudget == 0 {
		// gas budget 0 means its a view call, so we give it max gas and tokens
		return reqctx.vm.chainInfo.GasLimits.MaxGasExternalViewCall, math.MaxUint64
	}

	// make sure the gasBuget is at least >= than the allowed minimum
	if gasBudget < reqctx.vm.chainInfo.GasLimits.MinGasPerRequest {
		gasBudget = reqctx.vm.chainInfo.GasLimits.MinGasPerRequest
	}

	// calculate how many tokens for gas fee can be guaranteed after taking into account the allowance
	guaranteedFeeTokens := reqctx.calcGuaranteedFeeTokens()
	// calculate how many tokens maximum will be charged taking into account the budget
	f1, f2 := reqctx.vm.chainInfo.GasFeePolicy.FeeFromGasBurned(gasBudget, guaranteedFeeTokens)
	maxTokensToSpendForGasFee = f1 + f2
	// calculate affordableGas gas budget
	affordableGas := reqctx.vm.chainInfo.GasFeePolicy.GasBudgetFromTokens(guaranteedFeeTokens)
	// adjust gas budget to what is affordable
	affordableGas = util.MinUint64(gasBudget, affordableGas)
	// cap gas to the maximum allowed per tx
	return util.MinUint64(affordableGas, reqctx.vm.chainInfo.GasLimits.MaxGasPerRequest), maxTokensToSpendForGasFee
}

// calcGuaranteedFeeTokens return the maximum tokens (base tokens or native) can be guaranteed for the fee,
// taking into account allowance (which must be 'reserved')
func (reqctx *requestContext) calcGuaranteedFeeTokens() uint64 {
	tokensGuaranteed := reqctx.GetBaseTokensBalance(reqctx.req.SenderAccount())
	// safely subtract the allowed from the sender to the target
	if allowed := reqctx.req.Allowance(); allowed != nil {
		if tokensGuaranteed < allowed.BaseTokens {
			tokensGuaranteed = 0
		} else {
			tokensGuaranteed -= allowed.BaseTokens
		}
	}
	return tokensGuaranteed
}

// chargeGasFee takes burned tokens from the sender's account
// It should always be enough because gas budget is set affordable
func (reqctx *requestContext) chargeGasFee() {
	defer func() {
		// add current request gas burn to the total of the block
		reqctx.vm.blockGas.burned += reqctx.gas.burned
	}()

	// ensure at least the minimum amount of gas is charged
	minGas := gas.BurnCodeMinimumGasPerRequest1P.Cost(0)
	if reqctx.gas.burned < minGas {
		reqctx.gas.burned = minGas
	}

	if !reqctx.shouldChargeGasFee() {
		return
	}

	availableToPayFee := reqctx.gas.maxTokensToSpendForGasFee
	if !reqctx.vm.task.EstimateGasMode && !reqctx.vm.chainInfo.GasFeePolicy.IsEnoughForMinimumFee(availableToPayFee) {
		// user didn't specify enough base tokens to cover the minimum request fee, charge whatever is present in the user's account
		availableToPayFee = reqctx.GetSenderTokenBalanceForFees()
	}

	// total fees to charge
	sendToPayout, sendToValidator := reqctx.vm.chainInfo.GasFeePolicy.FeeFromGasBurned(reqctx.GasBurned(), availableToPayFee)
	reqctx.gas.feeCharged = sendToPayout + sendToValidator

	// calc gas totals
	reqctx.vm.blockGas.feeCharged += reqctx.gas.feeCharged

	if reqctx.vm.task.EstimateGasMode {
		// If estimating gas, compute the gas fee but do not attempt to charge
		return
	}

	sender := reqctx.req.SenderAccount()
	if sendToValidator != 0 {
		transferToValidator := &isc.Assets{}
		transferToValidator.BaseTokens = sendToValidator
		mustMoveBetweenAccounts(reqctx.uncommittedState, sender, reqctx.vm.task.ValidatorFeeTarget, transferToValidator)
	}

	// ensure common account has at least minBalanceInCommonAccount, and transfer the rest of gas fee to payout AgentID
	// if the payout AgentID is not set in governance contract, then chain owner will be used
	var minBalanceInCommonAccount uint64
	withContractState(reqctx.uncommittedState, governance.Contract, func(s kv.KVStore) {
		minBalanceInCommonAccount = governance.MustGetMinCommonAccountBalance(s)
	})
	commonAccountBal := reqctx.GetBaseTokensBalance(accounts.CommonAccount())
	if commonAccountBal < minBalanceInCommonAccount {
		// pay to common account since the balance of common account is less than minSD
		transferToCommonAcc := sendToPayout
		sendToPayout = 0
		if commonAccountBal+transferToCommonAcc > minBalanceInCommonAccount {
			excess := (commonAccountBal + transferToCommonAcc) - minBalanceInCommonAccount
			transferToCommonAcc -= excess
			sendToPayout = excess
		}
		mustMoveBetweenAccounts(reqctx.uncommittedState, sender, accounts.CommonAccount(), isc.NewAssetsBaseTokens(transferToCommonAcc))
	}
	if sendToPayout > 0 {
		payoutAgentID := reqctx.vm.payoutAgentID()
		mustMoveBetweenAccounts(reqctx.uncommittedState, sender, payoutAgentID, isc.NewAssetsBaseTokens(sendToPayout))
	}
}

func (reqctx *requestContext) LocateProgram(programHash hashing.HashValue) (vmtype string, binary []byte, err error) {
	return reqctx.vm.locateProgram(reqctx.chainStateWithGasBurn(), programHash)
}

func (reqctx *requestContext) Processors() *processors.Cache {
	return reqctx.vm.task.Processors
}

func (reqctx *requestContext) GetContractRecord(contractHname isc.Hname) (ret *root.ContractRecord) {
	ret = findContractByHname(reqctx.chainStateWithGasBurn(), contractHname)
	if ret == nil {
		reqctx.GasBurn(gas.BurnCodeCallTargetNotFound)
		panic(vm.ErrContractNotFound.Create(contractHname))
	}
	return ret
}

func (vmctx *vmContext) loadChainConfig() {
	vmctx.chainInfo = governance.NewStateAccess(vmctx.stateDraft).ChainInfo(vmctx.ChainID())
}

// mustCheckTransactionSize panics with ErrMaxTransactionSizeExceeded if the estimated transaction size exceeds the limit
func (vmctx *vmContext) mustCheckTransactionSize() {
	essence, _ := vmctx.BuildTransactionEssence(state.L1CommitmentNil, false)
	tx := transaction.MakeAnchorTransaction(essence, &iotago.Ed25519Signature{})
	if tx.Size() > parameters.L1().MaxPayloadSize {
		panic(vmexceptions.ErrMaxTransactionSizeExceeded)
	}
}