// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
)

func makeGasSStoreFunc(clearingRefund uint64) gasFunc {
	return func(evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		// If we fail the minimum gas availability invariant, fail (0)
		if contract.Gas <= params.SstoreSentryGasEIP2200 {
			return 0, errors.New("not enough gas for reentrancy sentry")
		}
		// Gas sentry honoured, do the actual gas calculation based on the stored value
		var (
			y, x    = stack.Back(1), stack.Peek()
			slot    = common.Hash(x.Bytes32())
			current = evm.StateDB.GetState(contract.Address(), slot)
			cost    = uint64(0)
		)
		// Check slot presence in the access list
		if addrPresent, slotPresent := evm.StateDB.SlotInAccessList(contract.Address(), slot); !slotPresent {
			cost = params.ColdSloadCostEIP2929
			// If the caller cannot afford the cost, this change will be rolled back
			evm.StateDB.AddSlotToAccessList(contract.Address(), slot)
			if !addrPresent {
				// Once we're done with YOLOv2 and schedule this for mainnet, might
				// be good to remove this panic here, which is just really a
				// canary to have during testing
				panic("impossible case: address was not present in access list during sstore op")
			}
		}
		value := common.Hash(y.Bytes32())

		if current == value { // noop (1)
			// EIP 2200 original clause:
			//		return params.SloadGasEIP2200, nil
			return cost + params.WarmStorageReadCostEIP2929, nil // SLOAD_GAS
		}
		original := evm.StateDB.GetCommittedState(contract.Address(), x.Bytes32())
		if original == current {
			if original == (common.Hash{}) { // create slot (2.1.1)
				return cost + params.SstoreSetGasEIP2200, nil
			}
			if value == (common.Hash{}) { // delete slot (2.1.2b)
				evm.StateDB.AddRefund(clearingRefund)
			}
			// EIP-2200 original clause:
			//		return params.SstoreResetGasEIP2200, nil // write existing slot (2.1.2)
			return cost + (params.SstoreResetGasEIP2200 - params.ColdSloadCostEIP2929), nil // write existing slot (2.1.2)
		}
		if original != (common.Hash{}) {
			if current == (common.Hash{}) { // recreate slot (2.2.1.1)
				evm.StateDB.SubRefund(clearingRefund)
			} else if value == (common.Hash{}) { // delete slot (2.2.1.2)
				evm.StateDB.AddRefund(clearingRefund)
			}
		}
		if original == value {
			if original == (common.Hash{}) { // reset to original inexistent slot (2.2.2.1)
				// EIP 2200 Original clause:
				// evm.StateDB.AddRefund(params.SstoreSetGasEIP2200 - params.SloadGasEIP2200)
				evm.StateDB.AddRefund(params.SstoreSetGasEIP2200 - params.WarmStorageReadCostEIP2929)
			} else { // reset to original existing slot (2.2.2.2)
				// EIP 2200 Original clause:
				//	evm.StateDB.AddRefund(params.SstoreResetGasEIP2200 - params.SloadGasEIP2200)
				// - SSTORE_RESET_GAS redefined as (5000 - COLD_SLOAD_COST)
				// - SLOAD_GAS redefined as WARM_STORAGE_READ_COST
				// Final: (5000 - COLD_SLOAD_COST) - WARM_STORAGE_READ_COST
				evm.StateDB.AddRefund((params.SstoreResetGasEIP2200 - params.ColdSloadCostEIP2929) - params.WarmStorageReadCostEIP2929)
			}
		}
		// EIP-2200 original clause:
		// return params.SloadGasEIP2200, nil // dirty update (2.2)
		return cost + params.WarmStorageReadCostEIP2929, nil // dirty update (2.2)
	}
}

// gasSLoadEIP2929 calculates dynamic gas for SLOAD according to EIP-2929
// For SLOAD, if the (address, storage_key) pair (where address is the address of the contract
// whose storage is being read) is not yet in accessed_storage_keys,
// charge 2100 gas and add the pair to accessed_storage_keys.
// If the pair is already in accessed_storage_keys, charge 100 gas.
func gasSLoadEIP2929(evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	loc := stack.Peek()
	slot := common.Hash(loc.Bytes32())
	// Check slot presence in the access list
	if _, slotPresent := evm.StateDB.SlotInAccessList(contract.Address(), slot); !slotPresent {
		// If the caller cannot afford the cost, this change will be rolled back
		// If he does afford it, we can skip checking the same thing later on, during execution
		evm.StateDB.AddSlotToAccessList(contract.Address(), slot)
		return params.ColdSloadCostEIP2929, nil
	}
	return params.WarmStorageReadCostEIP2929, nil
}

// gasExtCodeCopyEIP2929 implements extcodecopy according to EIP-2929
// EIP spec:
// > If the target is not in accessed_addresses,
// > charge COLD_ACCOUNT_ACCESS_COST gas, and add the address to accessed_addresses.
// > Otherwise, charge WARM_STORAGE_READ_COST gas.
func gasExtCodeCopyEIP2929(evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	// memory expansion first (dynamic part of pre-2929 implementation)
	gas, err := gasExtCodeCopy(evm, contract, stack, mem, memorySize)
	if err != nil {
		return 0, err
	}
	addr := common.Address(stack.Peek().Bytes20())
	// Check slot presence in the access list
	if !evm.StateDB.AddressInAccessList(addr) {
		evm.StateDB.AddAddressToAccessList(addr)
		var overflow bool
		// We charge (cold-warm), since 'warm' is already charged as constantGas
		if gas, overflow = math.SafeAdd(gas, params.ColdAccountAccessCostEIP2929-params.WarmStorageReadCostEIP2929); overflow {
			return 0, ErrGasUintOverflow
		}
		return gas, nil
	}
	return gas, nil
}

// gasEip2929AccountCheck checks whether the first stack item (as address) is present in the access list.
// If it is, this method returns '0', otherwise 'cold-warm' gas, presuming that the opcode using it
// is also using 'warm' as constant factor.
// This method is used by:
// - extcodehash,
// - extcodesize,
// - (ext) balance
func gasEip2929AccountCheck(evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	addr := common.Address(stack.Peek().Bytes20())
	// Check slot presence in the access list
	if !evm.StateDB.AddressInAccessList(addr) {
		// If the caller cannot afford the cost, this change will be rolled back
		evm.StateDB.AddAddressToAccessList(addr)
		// The warm storage read cost is already charged as constantGas
		return params.ColdAccountAccessCostEIP2929 - params.WarmStorageReadCostEIP2929, nil
	}
	return 0, nil
}

func makeCallVariantGasCallEIP2929(oldCalculator gasFunc) gasFunc {
	return func(evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		addr := common.Address(stack.Back(1).Bytes20())
		// Check slot presence in the access list
		warmAccess := evm.StateDB.AddressInAccessList(addr)
		// The WarmStorageReadCostEIP2929 (100) is already deducted in the form of a constant cost, so
		// the cost to charge for cold access, if any, is Cold - Warm
		coldCost := params.ColdAccountAccessCostEIP2929 - params.WarmStorageReadCostEIP2929
		if !warmAccess {
			evm.StateDB.AddAddressToAccessList(addr)
			// Charge the remaining difference here already, to correctly calculate available
			// gas for call
			if !contract.UseGas(coldCost) {
				return 0, ErrOutOfGas
			}
		}
		// Now call the old calculator, which takes into account
		// - create new account
		// - transfer value
		// - memory expansion
		// - 63/64ths rule
		gas, err := oldCalculator(evm, contract, stack, mem, memorySize)
		if warmAccess || err != nil {
			return gas, err
		}
		// In case of a cold access, we temporarily add the cold charge back, and also
		// add it to the returned gas. By adding it to the return, it will be charged
		// outside of this function, as part of the dynamic gas, and that will make it
		// also become correctly reported to tracers.
		contract.Gas += coldCost
		return gas + coldCost, nil
	}
}

var (
	gasCallEIP2929         = makeCallVariantGasCallEIP2929(gasCall)
	gasDelegateCallEIP2929 = makeCallVariantGasCallEIP2929(gasDelegateCall)
	gasStaticCallEIP2929   = makeCallVariantGasCallEIP2929(gasStaticCall)
	gasCallCodeEIP2929     = makeCallVariantGasCallEIP2929(gasCallCode)
	gasSelfdestructEIP2929 = makeSelfdestructGasFn(true)
	// gasSelfdestructEIP3529 implements the changes in EIP-2539 (no refunds)
	gasSelfdestructEIP3529 = makeSelfdestructGasFn(false)

	// gasSStoreEIP2929 implements gas cost for SSTORE according to EIP-2929
	//
	// When calling SSTORE, check if the (address, storage_key) pair is in accessed_storage_keys.
	// If it is not, charge an additional COLD_SLOAD_COST gas, and add the pair to accessed_storage_keys.
	// Additionally, modify the parameters defined in EIP 2200 as follows:
	//
	// Parameter 	Old value 	New value
	// SLOAD_GAS 	800 	= WARM_STORAGE_READ_COST
	// SSTORE_RESET_GAS 	5000 	5000 - COLD_SLOAD_COST
	//
	//The other parameters defined in EIP 2200 are unchanged.
	// see gasSStoreEIP2200(...) in core/vm/gas_table.go for more info about how EIP 2200 is specified
	gasSStoreEIP2929 = makeGasSStoreFunc(params.SstoreClearsScheduleRefundEIP2200)

	// gasSStoreEIP2539 implements gas cost for SSTORE according to EIP-2539
	// Replace `SSTORE_CLEARS_SCHEDULE` with `SSTORE_RESET_GAS + ACCESS_LIST_STORAGE_KEY_COST` (4,800)
	gasSStoreEIP3529 = makeGasSStoreFunc(params.SstoreClearsScheduleRefundEIP3529)
)

// makeSelfdestructGasFn can create the selfdestruct dynamic gas function for EIP-2929 and EIP-2539
func makeSelfdestructGasFn(refundsEnabled bool) gasFunc {
	gasFunc := func(evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		var (
			gas     uint64
			address = common.Address(stack.Peek().Bytes20())
		)
		if !evm.StateDB.AddressInAccessList(address) {
			// If the caller cannot afford the cost, this change will be rolled back
			evm.StateDB.AddAddressToAccessList(address)
			gas = params.ColdAccountAccessCostEIP2929
		}
		// if empty and transfers value
		if evm.StateDB.Empty(address) && evm.StateDB.GetBalance(contract.Address()).Sign() != 0 {
			gas += params.CreateBySelfdestructGas
		}
		if refundsEnabled && !evm.StateDB.HasSuicided(contract.Address()) {
			evm.StateDB.AddRefund(params.SelfdestructRefundGas)
		}
		return gas, nil
	}
	return gasFunc
}

func gasMCopy(
	evm *EVM,
	contract *Contract,
	stack *Stack,
	mem *Memory,
	memorySize uint64,
) (uint64, error) {
	// The MCOPY opcode is expected to pop 3 stack items in the execution function:
	//   [length, srcOffset, destOffset]
	// Typically the top of stack is length, then srcOffset, then destOffset,
	// but we read them in reverse order with 'Back()'.
	length := stack.Back(0).Uint64()
	src := stack.Back(1).Uint64()
	dst := stack.Back(2).Uint64()

	// --- 1) Calculate memory expansion costs ---
	// For the copy to be valid, we need memory to cover [dst..dst+length-1]
	// and [src..src+length-1].
	endDst, overflow1 := math.SafeAdd(dst, length)
	endSrc, overflow2 := math.SafeAdd(src, length)
	if overflow1 || overflow2 {
		fmt.Printf("MCOPY: gas Over flow 1\n")
		return 0, ErrGasUintOverflow
	}
	maxEnd := endDst
	if endSrc > maxEnd {
		maxEnd = endSrc
	}
	// The memoryGasCost function is used to see if we need to expand memory
	// beyond 'memorySize' up to 'maxEnd'.
	memGas, err := memoryGasCost(mem, maxEnd)
	if err != nil {
		fmt.Printf("MCOPY: gas Over flow 2\n")
		return 0, err
	}

	// --- 2) Calculate the "per-byte" copy cost ---
	// EIP-5656 suggests 3 gas per byte (similar to other copy ops).
	const copyGasPerByte = 3
	copyCost, overflow3 := math.SafeMul(copyGasPerByte, length)
	if overflow3 {
		fmt.Printf("MCOPY: gas Over flow 3\n")
		return 0, ErrGasUintOverflow
	}

	// Combine memory expansion + copy cost
	totalGas, overflow4 := math.SafeAdd(memGas, copyCost)
	if overflow4 {
		fmt.Printf("MCOPY: gas Over flow 4\n")
		return 0, ErrGasUintOverflow
	}
	fmt.Printf("MCOPY: endSrc=%d, endDst=%d, maxEnd=%d\n", endSrc, endDst, maxEnd)
	return totalGas, nil
}

var ErrMaxInitCodeSizeExceeded = errors.New("init code size exceeds maximum allowed by EIP-3860")

// gasCreateEIP3860 calculates the additional dynamic gas cost for the CREATE opcode according to EIP-3860.
// It assumes that the top two items on the stack are:
//   - initcodeSize (top of stack)
//   - initcodeOffset (next item)
//
// and that the constant base cost is already applied separately.
func gasCreateEIP3860(
	evm *EVM,
	contract *Contract,
	stack *Stack,
	mem *Memory,
	memorySize uint64,
) (uint64, error) {
	// -----------------------------------------------------------
	// 1. Read the initcode offset and size from the stack.
	// For CREATE, the typical stack layout is: [ value, initcodeOffset, initcodeSize ]
	// where initcodeSize is at the top of the stack.
	size := stack.Back(0).Uint64()
	offset := stack.Back(1).Uint64()

	// -----------------------------------------------------------
	// 2. Enforce the EIP-3860 maximum initcode size limit.
	const maxInitCodeSize = 49152
	if size > maxInitCodeSize {
		return 0, ErrMaxInitCodeSizeExceeded
	}

	// -----------------------------------------------------------
	// 3. Calculate the required memory expansion.
	// The memory must cover the range [offset, offset+size).
	requiredMemSize, overflow := math.SafeAdd(offset, size)
	if overflow {
		return 0, ErrGasUintOverflow
	}
	memGas, err := memoryGasCost(mem, requiredMemSize)
	if err != nil {
		return 0, err
	}

	// -----------------------------------------------------------
	// 4. Compute the EIP-3860 overhead: 2 gas per 32-byte chunk of initcode.
	// That is, overhead = 2 * ceil(size / 32) = 2 * ((size + 31) / 32)
	chunkCount := (size + 31) / 32
	overhead, overflow2 := math.SafeMul(chunkCount, 2)
	if overflow2 {
		return 0, ErrGasUintOverflow
	}

	// -----------------------------------------------------------
	// 5. Combine the memory expansion cost and the EIP-3860 overhead.
	totalGas, overflow3 := math.SafeAdd(memGas, overhead)
	if overflow3 {
		return 0, ErrGasUintOverflow
	}
	return totalGas, nil
}

// gasCreate2EIP3860 calculates the extra dynamic gas cost for the CREATE2 opcode according to EIP-3860.
// It assumes that on the stack, the top item is the initcode size and the next item is the initcode offset.
func gasCreate2EIP3860(
	evm *EVM,
	contract *Contract,
	stack *Stack,
	mem *Memory,
	memorySize uint64,
) (uint64, error) {
	// -----------------------------------------------------------
	// 1. Read the initcode offset and size from the stack.
	// Assumption: For CREATE2, the stack layout is:
	//   ... , initcode_offset, initcode_size
	// with initcode_size at the top of the stack.
	size := stack.Back(0).Uint64()
	offset := stack.Back(1).Uint64()

	// -----------------------------------------------------------
	// 2. Enforce maximum initcode size per EIP-3860.
	const maxInitCodeSize = 49152
	if size > maxInitCodeSize {
		return 0, ErrMaxInitCodeSizeExceeded
	}

	// -----------------------------------------------------------
	// 3. Calculate the required memory expansion.
	// The memory must cover [offset, offset+size). Use SafeAdd to prevent overflows.
	requiredMemSize, overflow := math.SafeAdd(offset, size)
	if overflow {
		return 0, ErrGasUintOverflow
	}
	memGas, err := memoryGasCost(mem, requiredMemSize)
	if err != nil {
		return 0, err
	}

	// -----------------------------------------------------------
	// 4. Compute the EIP-3860 overhead.
	// EIP-3860 charges 2 gas for every 32-byte chunk of initcode.
	chunkCount := (size + 31) / 32 // rounds up to the nearest 32 bytes
	overhead, overflow := math.SafeMul(chunkCount, 2)
	if overflow {
		return 0, ErrGasUintOverflow
	}

	// -----------------------------------------------------------
	// 5. Combine the memory expansion cost and the EIP-3860 overhead.
	totalGas, overflow := math.SafeAdd(memGas, overhead)
	if overflow {
		return 0, ErrGasUintOverflow
	}
	return totalGas, nil
}
