package core

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

func TestStateTransition_EnablingExcessGasChargingEnablesExcessGasCharging(t *testing.T) {
	msg := &Message{
		From:     common.Address{12},
		GasLimit: 100_000,
	}

	config := vm.Config{
		ChargeExcessGas: false,
	}

	resultWithoutCharge, err := runTestTransaction(msg, config)
	if err != nil {
		t.Fatalf("Error running transaction: %v", err)
	}

	config.ChargeExcessGas = true
	resultWithCharge, err := runTestTransaction(msg, config)
	if err != nil {
		t.Fatalf("Error running transaction: %v", err)
	}

	// When enabled, 10% of the excess gas is charged
	want := (msg.GasLimit - resultWithoutCharge.UsedGas) / 10
	diff := resultWithCharge.UsedGas - resultWithoutCharge.UsedGas
	if diff != want {
		t.Fatalf("Expected difference in gas usage to be %d, got %d", want, diff)
	}
}

func TestStateTransition_ExcessiveGasChargesAreIgnoredForTheZeroSender(t *testing.T) {
	msg := &Message{
		From:     common.Address{0},
		GasLimit: 100_000,
	}

	config := vm.Config{
		ChargeExcessGas: false,
	}

	resultWithoutCharge, err := runTestTransaction(msg, config)
	if err != nil {
		t.Fatalf("Error running transaction: %v", err)
	}

	config.ChargeExcessGas = true
	resultWithCharge, err := runTestTransaction(msg, config)
	if err != nil {
		t.Fatalf("Error running transaction: %v", err)
	}

	if resultWithCharge.UsedGas != resultWithoutCharge.UsedGas {
		t.Fatalf("Expected gas usage to be the same, got %d and %d", resultWithoutCharge.UsedGas, resultWithCharge.UsedGas)
	}
}

func TestStateTransition_InsufficientBalanceCheckCanBeDisabled(t *testing.T) {
	msg := &Message{
		Value:    big.NewInt(2_000_000), // - all accounts have 1_000_000 in the dummy DB
		GasLimit: 100_000,
	}

	config := vm.Config{
		InsufficientBalanceIsNotAnError: false,
	}

	_, err := runTestTransaction(msg, config)
	if err == nil {
		t.Errorf("Expected error running transaction with not enough balance")
	}

	// When disabled, the transaction is still processed but reverted.
	config.InsufficientBalanceIsNotAnError = true
	result, err := runTestTransaction(msg, config)
	if err != nil {
		t.Errorf("Error running transaction: %v", err)
	}
	if errors.Is(result.Err, vm.ErrExecutionReverted) {
		t.Errorf("Expected error to be reverted, got %v", result.Err)
	}
}

func TestStateTransition_GasFeeCapCanBeIgnored(t *testing.T) {
	msg := &Message{
		GasLimit:  100_000,
		GasFeeCap: big.NewInt(8),  // 1M in each account should be enough
		GasPrice:  big.NewInt(12), // 1M is not enough for this gas price
	}

	config := vm.Config{
		IgnoreGasFeeCap: false,
	}

	// By default, the gas fee cap is enforced and the transaction should pass.
	_, err := runTestTransaction(msg, config)
	if err != nil {
		t.Errorf("expected transaction to pass when enforcing gas fee cap")
	}

	// When ignoring the gas fee cap, the transaction should be too expensive.
	config.IgnoreGasFeeCap = true
	_, err = runTestTransaction(msg, config)
	if err == nil {
		t.Errorf("expected transaction to fail when ignoring gas fee cap")
	}
}

func TestStateTransition_GasTipPaymentToCoinbaseCanBeSkipped(t *testing.T) {
	msg := &Message{
		GasLimit: 100_000,
		GasPrice: big.NewInt(5),
	}

	config := vm.Config{
		SkipTipPaymentToCoinbase: false,
	}

	// By default, gas fees are send to the coinbase.
	result, balance, err := runTestTransactionAndGetBalance(msg, config, testCoinbase)
	if err != nil {
		t.Errorf("transaction failed: %v", err)
	}
	if got, want := balance.Uint64(), uint64(1_000_000+result.UsedGas*5); got != want {
		t.Errorf("expected coinbase balance to be %d, got %d", want, got)
	}

	// When disabled, transfers are skipped.
	config.SkipTipPaymentToCoinbase = true
	_, balance, err = runTestTransactionAndGetBalance(msg, config, testCoinbase)
	if err != nil {
		t.Errorf("transaction failed: %v", err)
	}
	if got, want := balance.Uint64(), uint64(1_000_000); got != want {
		t.Errorf("expected coinbase balance to be %d, got %d", want, got)
	}
}

///////////////////////////
// Helper functions

var testCoinbase = common.Address{15}

func runTestTransaction(msg *Message, config vm.Config) (*ExecutionResult, error) {
	result, _, err := runTestTransactionAndGetBalance(msg, config, common.Address{})
	return result, err
}

func runTestTransactionAndGetBalance(
	msg *Message,
	config vm.Config,
	address common.Address,
) (*ExecutionResult, *uint256.Int, error) {
	evm := vm.NewEVM(
		vm.BlockContext{
			Transfer:    func(vm.StateDB, common.Address, common.Address, *uint256.Int) {},
			CanTransfer: func(vm.StateDB, common.Address, *uint256.Int) bool { return true },
			Coinbase:    testCoinbase,
		},
		vm.TxContext{},
		&dummyStateDB{
			accountToTrack: address,
		},
		&params.ChainConfig{},
		config,
	)

	if msg.To == nil {
		msg.To = &common.Address{14}
	}
	if msg.GasPrice == nil {
		msg.GasPrice = big.NewInt(0)
	}
	if msg.GasFeeCap == nil {
		msg.GasFeeCap = big.NewInt(0)
	}
	if msg.GasTipCap == nil {
		msg.GasTipCap = big.NewInt(0)
	}
	if msg.Value == nil {
		msg.Value = big.NewInt(0)
	}

	pool := new(GasPool)
	pool.AddGas(100_000)

	result, err := ApplyMessage(evm, msg, pool)
	return result, evm.StateDB.GetBalance(address), err
}

///////////////////////////
// dummyStateDB

type dummyStateDB struct {
	vm.StateDB

	accountToTrack common.Address
	accountBalance *uint256.Int
}

func (*dummyStateDB) Exist(common.Address) bool {
	return true
}

func (db *dummyStateDB) GetBalance(addr common.Address) *uint256.Int {
	if addr == db.accountToTrack && db.accountBalance != nil {
		return db.accountBalance
	}
	return uint256.NewInt(1_000_000)
}

func (db *dummyStateDB) AddBalance(addr common.Address, value *uint256.Int, _ tracing.BalanceChangeReason) {
	if addr == db.accountToTrack {
		cur := db.GetBalance(addr)
		db.accountBalance = cur.Add(cur, value)
	}
}

func (*dummyStateDB) SubBalance(common.Address, *uint256.Int, tracing.BalanceChangeReason) {
	// ignored
}

func (*dummyStateDB) GetNonce(common.Address) uint64 {
	return 0
}

func (*dummyStateDB) SetNonce(common.Address, uint64) {
	// ignored
}

func (*dummyStateDB) GetCodeHash(common.Address) common.Hash {
	return common.Hash{}
}

func (*dummyStateDB) GetCode(common.Address) []byte {
	return nil
}

func (*dummyStateDB) AddRefund(uint64) {
	// ignored
}
func (*dummyStateDB) SubRefund(uint64) {
	// ignored
}
func (*dummyStateDB) GetRefund() uint64 {
	return 0
}

func (*dummyStateDB) Prepare(params.Rules, common.Address, common.Address, *common.Address, []common.Address, types.AccessList) {
	// ignored
}

func (*dummyStateDB) Snapshot() int {
	return 0
}

func (*dummyStateDB) RevertToSnapshot(int) {
	// ignored
}

func (*dummyStateDB) Witness() *stateless.Witness {
	return nil
}
