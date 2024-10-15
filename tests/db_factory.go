package tests

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
)

// TestStateDB allows for switching database implementation for running tests.
// It is an extension of vm.StateDB with additional methods that were originally
// available only at an implementation level of the geth database.
// Not all methods have to be available in all implementations, and clients
// should pair expected method outputs with the actual implementation.
type TestStateDB interface {
	vm.StateDB

	// Database returns the underlying database.
	Database() state.Database

	// Logs returns the logs of the current transaction.
	Logs() []*types.Log

	SetLogger(l *tracing.Hooks)

	// SetBalance sets the balance of the given account.
	SetBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason)

	// IntermediateRoot returns current state root hash.
	IntermediateRoot(deleteEmptyObjects bool) common.Hash

	// Commit commits the state to the underlying trie database and returns state root hash.
	Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error)
}

// TestContextFactory is an interface for creating test configurations.
type TestContextFactory interface {

	// NewTestStateDB creates a new StateTestState instance.
	NewTestStateDB(accounts types.GenesisAlloc) StateTestState
}
