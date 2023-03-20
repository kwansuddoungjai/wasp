// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chains

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smGPAUtils"
	"github.com/iotaledger/wasp/packages/chains/accessMgr"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/shutdown"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type Provider func() *Chains // TODO: Use DI instead of that.

func (chains Provider) ChainProvider() func(chainID isc.ChainID) chain.Chain {
	return func(chainID isc.ChainID) chain.Chain {
		return chains().Get(chainID)
	}
}

type ChainProvider func(chainID isc.ChainID) chain.Chain

type Chains struct {
	ctx                              context.Context
	log                              *logger.Logger
	nodeConnection                   chain.NodeConnection
	processorConfig                  *processors.Config
	offledgerBroadcastUpToNPeers     int
	offledgerBroadcastInterval       time.Duration
	pullMissingRequestsFromCommittee bool
	networkProvider                  peering.NetworkProvider
	trustedNetworkManager            peering.TrustedNetworkManager
	trustedNetworkListenerCancel     context.CancelFunc
	chainStateStoreProvider          database.ChainStateKVStoreProvider
	walEnabled                       bool
	walFolderPath                    string

	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	dkShareRegistryProvider     registry.DKShareRegistryProvider
	nodeIdentityProvider        registry.NodeIdentityProvider
	consensusStateRegistry      cmtLog.ConsensusStateRegistry
	chainListener               chain.ChainListener

	mutex     sync.RWMutex
	allChains map[isc.ChainID]*activeChain
	accessMgr accessMgr.AccessMgr

	cleanupFunc context.CancelFunc

	shutdownCoordinator *shutdown.Coordinator
	chainMetrics        *metrics.ChainMetrics
	blockWALMetrics     *metrics.BlockWALMetrics
}

type activeChain struct {
	chain      chain.Chain
	cancelFunc context.CancelFunc
}

func New(
	log *logger.Logger,
	nodeConnection chain.NodeConnection,
	processorConfig *processors.Config,
	offledgerBroadcastUpToNPeers int, // TODO: Unused for now.
	offledgerBroadcastInterval time.Duration, // TODO: Unused for now.
	pullMissingRequestsFromCommittee bool, // TODO: Unused for now.
	networkProvider peering.NetworkProvider,
	trustedNetworkManager peering.TrustedNetworkManager,
	chainStateStoreProvider database.ChainStateKVStoreProvider,
	walEnabled bool,
	walFolderPath string,
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	dkShareRegistryProvider registry.DKShareRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	consensusStateRegistry cmtLog.ConsensusStateRegistry,
	chainListener chain.ChainListener,
	shutdownCoordinator *shutdown.Coordinator,
	chainMetrics *metrics.ChainMetrics,
	blockWALMetrics *metrics.BlockWALMetrics,
) *Chains {
	ret := &Chains{
		log:                              log,
		allChains:                        map[isc.ChainID]*activeChain{},
		nodeConnection:                   nodeConnection,
		processorConfig:                  processorConfig,
		offledgerBroadcastUpToNPeers:     offledgerBroadcastUpToNPeers,
		offledgerBroadcastInterval:       offledgerBroadcastInterval,
		pullMissingRequestsFromCommittee: pullMissingRequestsFromCommittee,
		networkProvider:                  networkProvider,
		trustedNetworkManager:            trustedNetworkManager,
		chainStateStoreProvider:          chainStateStoreProvider,
		walEnabled:                       walEnabled,
		walFolderPath:                    walFolderPath,
		chainRecordRegistryProvider:      chainRecordRegistryProvider,
		dkShareRegistryProvider:          dkShareRegistryProvider,
		nodeIdentityProvider:             nodeIdentityProvider,
		chainListener:                    nil, // See bellow.
		consensusStateRegistry:           consensusStateRegistry,
		shutdownCoordinator:              shutdownCoordinator,
		chainMetrics:                     chainMetrics,
		blockWALMetrics:                  blockWALMetrics,
	}
	ret.chainListener = NewChainsListener(chainListener, ret.chainAccessUpdatedCB)
	return ret
}

func (c *Chains) Run(ctx context.Context) error {
	if err := c.nodeConnection.WaitUntilInitiallySynced(ctx); err != nil {
		return fmt.Errorf("waiting for L1 node to become sync failed, error: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.ctx != nil {
		return errors.New("chains already running")
	}
	c.ctx = ctx

	c.accessMgr = accessMgr.New(ctx, c.chainServersUpdatedCB, c.nodeIdentityProvider.NodeIdentity(), c.networkProvider, c.log.Named("AM"))
	c.trustedNetworkListenerCancel = c.trustedNetworkManager.TrustedPeersListener(c.trustedPeersUpdatedCB)

	unhook := c.chainRecordRegistryProvider.Events().ChainRecordModified.Hook(func(event *registry.ChainRecordModifiedEvent) {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
		if chain, ok := c.allChains[event.ChainRecord.ChainID()]; ok {
			chain.chain.ConfigUpdated(event.ChainRecord.AccessNodes)
		}
	}).Unhook
	c.cleanupFunc = unhook

	return c.activateAllFromRegistry() //nolint:contextcheck
}

func (c *Chains) Close() {
	util.ExecuteIfNotNil(c.cleanupFunc)
	for _, c := range c.allChains {
		c.cancelFunc()
	}
	c.shutdownCoordinator.WaitNestedWithLogging(1 * time.Second)
	c.shutdownCoordinator.Done()
	util.ExecuteIfNotNil(c.trustedNetworkListenerCancel)
	c.trustedNetworkListenerCancel = nil
}

func (c *Chains) trustedPeersUpdatedCB(trustedPeers []*peering.TrustedPeer) {
	trustedPubKeys := lo.Map(trustedPeers, func(tp *peering.TrustedPeer) *cryptolib.PublicKey { return tp.PubKey() })
	c.accessMgr.TrustedNodes(trustedPubKeys)
}

func (c *Chains) chainServersUpdatedCB(chainID isc.ChainID, servers []*cryptolib.PublicKey) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	ch, ok := c.allChains[chainID]
	if !ok {
		return
	}
	ch.chain.ServersUpdated(servers)
}

func (c *Chains) chainAccessUpdatedCB(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey) {
	c.accessMgr.ChainAccessNodes(chainID, accessNodes)
}

func (c *Chains) activateAllFromRegistry() error {
	var innerErr error
	if err := c.chainRecordRegistryProvider.ForEachActiveChainRecord(func(chainRecord *registry.ChainRecord) bool {
		chainID := chainRecord.ChainID()
		if err := c.activateWithoutLocking(chainID); err != nil {
			innerErr = fmt.Errorf("cannot activate chain %s: %w", chainRecord.ChainID(), err)
			return false
		}

		return true
	}); err != nil {
		return err
	}

	return innerErr
}

// activateWithoutLocking activates a chain in the node.
func (c *Chains) activateWithoutLocking(chainID isc.ChainID) error {
	if c.ctx == nil {
		return errors.New("run chains first")
	}
	//
	// Check, maybe it is already running.
	if _, ok := c.allChains[chainID]; ok {
		c.log.Debugf("Chain %v = %v is already activated", chainID.ShortString(), chainID.String())
		return nil
	}
	//
	// Activate the chain in the persistent store, if it is not activated yet.
	chainRecord, err := c.chainRecordRegistryProvider.ChainRecord(chainID)
	if err != nil {
		return fmt.Errorf("cannot get chain record for %v: %w", chainID, err)
	}
	if !chainRecord.Active {
		if _, err2 := c.chainRecordRegistryProvider.ActivateChainRecord(chainID); err2 != nil {
			return fmt.Errorf("cannot activate chain: %w", err2)
		}
	}

	chainKVStore, err := c.chainStateStoreProvider(chainID)
	if err != nil {
		return fmt.Errorf("error when creating chain KV store: %w", err)
	}

	// Initialize WAL
	chainLog := c.log.Named(chainID.ShortString())
	var chainWAL smGPAUtils.BlockWAL
	if c.walEnabled {
		chainWAL, err = smGPAUtils.NewBlockWAL(chainLog.Named("WAL"), c.walFolderPath, chainID, metrics.NewBlockWALMetric(c.blockWALMetrics, chainID))
		if err != nil {
			panic(fmt.Errorf("cannot create WAL: %w", err))
		}
	} else {
		chainWAL = smGPAUtils.NewEmptyBlockWAL()
	}

	chainCtx, chainCancel := context.WithCancel(c.ctx)
	newChain, err := chain.New(
		chainCtx,
		chainID,
		state.NewStore(chainKVStore),
		c.nodeConnection,
		c.nodeIdentityProvider.NodeIdentity(),
		c.processorConfig,
		c.dkShareRegistryProvider,
		c.consensusStateRegistry,
		chainWAL,
		c.chainListener,
		chainRecord.AccessNodes,
		c.networkProvider,
		metrics.NewChainMetric(c.chainMetrics, chainID),
		c.shutdownCoordinator.Nested(fmt.Sprintf("Chain-%s", chainID.AsAddress().String())),
		chainLog,
	)
	if err != nil {
		chainCancel()
		return fmt.Errorf("Chains.Activate: failed to create chain object: %w", err)
	}
	c.allChains[chainID] = &activeChain{
		chain:      newChain,
		cancelFunc: chainCancel,
	}

	c.log.Infof("activated chain: %v = %s", chainID.ShortString(), chainID.String())
	return nil
}

// Activate activates a chain in the node.
func (c *Chains) Activate(chainID isc.ChainID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.activateWithoutLocking(chainID)
}

// Deactivate a chain in the node.
func (c *Chains) Deactivate(chainID isc.ChainID) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, err := c.chainRecordRegistryProvider.DeactivateChainRecord(chainID); err != nil {
		return fmt.Errorf("cannot deactivate chain %v: %w", chainID, err)
	}

	ch, ok := c.allChains[chainID]
	if !ok {
		c.log.Debugf("chain is not active: %v = %s", chainID.ShortString(), chainID.String())
		return nil
	}
	ch.cancelFunc()
	c.accessMgr.ChainDismissed(chainID)
	delete(c.allChains, chainID)

	c.log.Debugf("chain has been deactivated: %v = %s", chainID.ShortString(), chainID.String())
	return nil
}

// Get returns active chain object or nil if it doesn't exist
// lazy unsubscribing
func (c *Chains) Get(chainID isc.ChainID) chain.Chain {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ret, ok := c.allChains[chainID]
	if !ok {
		return nil
	}
	return ret.chain
}

func (c *Chains) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return c.nodeConnection.GetMetrics()
}
