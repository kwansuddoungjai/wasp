package vmimpl

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *vmContext) runMigrations(chainState kv.KVStore, baseSchemaVersion uint32, allMigrations []migrations.Migration) {
	latestSchemaVersion := baseSchemaVersion + uint32(len(allMigrations))

	if vmctx.task.AnchorOutput.StateIndex == 0 {
		// initializing new chain -- set the schema to latest version
		withContractState(chainState, root.Contract, func(s kv.KVStore) {
			root.SetSchemaVersion(s, latestSchemaVersion)
		})
		return
	}

	var currentVersion uint32
	withContractState(chainState, root.Contract, func(s kv.KVStore) {
		currentVersion = root.GetSchemaVersion(s)
	})
	if currentVersion < baseSchemaVersion {
		panic(fmt.Sprintf("inconsistency: node with schema version %d is behind pruned migrations (should be >= %d)", currentVersion, baseSchemaVersion))
	}
	if currentVersion > latestSchemaVersion {
		panic(fmt.Sprintf("inconsistency: node with schema version %d is ahead latest schema version (should be <= %d)", currentVersion, latestSchemaVersion))
	}

	for currentVersion < latestSchemaVersion {
		migration := allMigrations[currentVersion-baseSchemaVersion]

		withContractState(chainState, migration.Contract, func(s kv.KVStore) {
			err := migration.Apply(s, vmctx.task.Log)
			if err != nil {
				panic(fmt.Sprintf("failed applying migration: %s", err))
			}
		})

		currentVersion++
		withContractState(chainState, root.Contract, func(s kv.KVStore) {
			root.SetSchemaVersion(s, currentVersion)
		})
	}
}