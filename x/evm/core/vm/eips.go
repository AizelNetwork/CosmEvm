// Copyright 2019 The go-ethereum Authors
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
	"sort"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var activators = map[string]func(*JumpTable){
	"ethereum_5656": enable5656,
	"ethereum_3855": enable3855,
	"ethereum_3529": enable3529,
	"ethereum_3198": enable3198,
	"ethereum_2929": enable2929,
	"ethereum_2200": enable2200,
	"ethereum_1884": enable1884,
	"ethereum_1344": enable1344,
}

// EnableEIP enables the given EIP on the config.
// This operation writes in-place, and callers need to ensure that the globally
// defined jump tables are not polluted.
func EnableEIP(eipName string, jt *JumpTable) error {
	enablerFn, ok := activators[eipName]
	if !ok {
		return fmt.Errorf("undefined eip %s", eipName)
	}
	enablerFn(jt)
	return nil
}

// ValidateEIPName checks if an EIP name is valid or not. The allowed
// name structure is a string that can be represented as "chainName" + "_" + "int".
func ValidateEIPName(eipName string) error {
	eipSplit := strings.Split(eipName, "_")
	if len(eipSplit) != 2 {
		return fmt.Errorf("eip name does not conform to structure 'chainName_Number'")
	}
	if _, err := strconv.Atoi(eipSplit[1]); err != nil {
		return fmt.Errorf("eip number should be convertible to int")
	}
	return nil
}

// ExistsEipActivator return true if the given EIP
// name is associated with an activator function.
// Return false otherwise.
func ExistsEipActivator(eipName string) bool {
	_, ok := activators[eipName]
	return ok
}

// ActivateableEips returns the sorted slice of EIP names
// that can be activated.
func ActivateableEips() []string {
	var names []string
	if len(activators) > 0 {
		for k := range activators {
			names = append(names, k)
		}
		sort.Strings(names)
	}
	return names
}

// enable1884 applies EIP-1884 to the given jump table:
// - Increase cost of BALANCE to 700
// - Increase cost of EXTCODEHASH to 700
// - Increase cost of SLOAD to 800
// - Define SELFBALANCE, with cost GasFastStep (5)
func enable1884(jt *JumpTable) {
	// Gas cost changes
	jt[SLOAD].constantGas = params.SloadGasEIP1884
	jt[BALANCE].constantGas = params.BalanceGasEIP1884
	jt[EXTCODEHASH].constantGas = params.ExtcodeHashGasEIP1884

	// New opcode
	jt[SELFBALANCE] = &operation{
		execute:     opSelfBalance,
		constantGas: GasFastStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

func opSelfBalance(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	balance, _ := uint256.FromBig(interpreter.evm.StateDB.GetBalance(scope.Contract.Address()))
	scope.Stack.Push(balance)
	return nil, nil
}

// enable1344 applies EIP-1344 (ChainID Opcode)
// - Adds an opcode that returns the current chain’s EIP-155 unique identifier
func enable1344(jt *JumpTable) {
	// New opcode
	jt[CHAINID] = &operation{
		execute:     opChainID,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

// opChainID implements CHAINID opcode
func opChainID(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	chainId, _ := uint256.FromBig(interpreter.evm.chainConfig.ChainID)
	scope.Stack.Push(chainId)
	return nil, nil
}

// enable2200 applies EIP-2200 (Rebalance net-metered SSTORE)
func enable2200(jt *JumpTable) {
	jt[SLOAD].constantGas = params.SloadGasEIP2200
	jt[SSTORE].dynamicGas = gasSStoreEIP2200
}

// enable2929 enables "EIP-2929: Gas cost increases for state access opcodes"
// https://eips.ethereum.org/EIPS/eip-2929
func enable2929(jt *JumpTable) {
	jt[SSTORE].dynamicGas = gasSStoreEIP2929

	jt[SLOAD].constantGas = 0
	jt[SLOAD].dynamicGas = gasSLoadEIP2929

	jt[EXTCODECOPY].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODECOPY].dynamicGas = gasExtCodeCopyEIP2929

	jt[EXTCODESIZE].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODESIZE].dynamicGas = gasEip2929AccountCheck

	jt[EXTCODEHASH].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODEHASH].dynamicGas = gasEip2929AccountCheck

	jt[BALANCE].constantGas = params.WarmStorageReadCostEIP2929
	jt[BALANCE].dynamicGas = gasEip2929AccountCheck

	jt[CALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[CALL].dynamicGas = gasCallEIP2929

	jt[CALLCODE].constantGas = params.WarmStorageReadCostEIP2929
	jt[CALLCODE].dynamicGas = gasCallCodeEIP2929

	jt[STATICCALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[STATICCALL].dynamicGas = gasStaticCallEIP2929

	jt[DELEGATECALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[DELEGATECALL].dynamicGas = gasDelegateCallEIP2929

	// This was previously part of the dynamic cost, but we're using it as a constantGas
	// factor here
	jt[SELFDESTRUCT].constantGas = params.SelfdestructGasEIP150
	jt[SELFDESTRUCT].dynamicGas = gasSelfdestructEIP2929
}

// enable3529 enabled "EIP-3529: Reduction in refunds":
// - Removes refunds for selfdestructs
// - Reduces refunds for SSTORE
// - Reduces max refunds to 20% gas
func enable3529(jt *JumpTable) {
	jt[SSTORE].dynamicGas = gasSStoreEIP3529
	jt[SELFDESTRUCT].dynamicGas = gasSelfdestructEIP3529
}

// enable3198 applies EIP-3198 (BASEFEE Opcode)
// - Adds an opcode that returns the current block's base fee.
func enable3198(jt *JumpTable) {
	// New opcode
	jt[BASEFEE] = &operation{
		execute:     opBaseFee,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

// opBaseFee implements BASEFEE opcode
func opBaseFee(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	baseFee, _ := uint256.FromBig(interpreter.evm.Context.BaseFee)
	scope.Stack.Push(baseFee)
	return nil, nil
}

// enable3855 applies EIP-3855 (PUSH0 opcode)
func enable3855(jt *JumpTable) {
	// New opcode
	jt[PUSH0] = &operation{
		execute:     opPush0,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

// opPush0 implements the PUSH0 opcode
func opPush0(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.Push(new(uint256.Int))
	return nil, nil
}

func enable5656(jt *JumpTable) {
	jt[MCOPY] = &operation{
		// This function will be called when EVM executes opcode 0x5E
		execute:     opMCopy,        // see next snippet
		dynamicGas:  gasMCopy,       // if you want a custom dynamic gas calc
		constantGas: 0,              // or GasQuickStep, etc. if you prefer
		minStack:    minStack(3, 0), // MCOPY pops 3 items (length, src, dst)
		maxStack:    maxStack(3, 0),
	}
}

var ErrMemoryOverflow = errors.New("memory overflow")

func opMCopy(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// Pop stack items (top of stack is length, then src, then dst).
	val := scope.Stack.Pop()  // Pop length
	length := (&val).Uint64() // Extract length

	val2 := scope.Stack.Pop() // Pop dst
	dst := (&val2).Uint64()   // Extract source offset

	val3 := scope.Stack.Pop() // Pop src
	src := (&val3).Uint64()   // Extract destination offset

	// If length == 0, no copying needed; just return.
	if length == 0 {
		return nil, nil
	}

	// Compute end offsets to see how far we need memory to extend
	endSrc, overflow1 := math.SafeAdd(src, length)
	endDst, overflow2 := math.SafeAdd(dst, length)
	if overflow1 || overflow2 {
		// Log the overflow details for debugging
		fmt.Printf("MCOPY: Memory overflow detected. Source end: %d, Destination end: %d\n", endSrc, endDst)
		return nil, ErrMemoryOverflow
	}

	// Resize memory so it covers [src..(src+length-1)] and [dst..(dst+length-1)].
	// We only need to resize to the largest of endSrc or endDst:
	maxEnd := endSrc
	if endDst > maxEnd {
		maxEnd = endDst
	}

	// Log the max end for memory resizing check
	fmt.Printf("MCOPY: Resizing memory to maxEnd: %d\n", maxEnd)
	scope.Memory.Resize(maxEnd)

	// Read from memory: get a pointer to the src segment
	srcData := scope.Memory.GetPtr(int64(src), int64(length))
	if srcData == nil {
		// Means offset is out of the actual memory store bounds
		fmt.Printf("MCOPY: Source memory access out of bounds at src=%d, length=%d\n", src, length)
		return nil, ErrMemoryOverflow
	}

	// Write to memory at [dst..dst+length]
	scope.Memory.Set(dst, length, srcData)

	// Log the details of the copy operation
	fmt.Printf("MCOPY: Successfully copied length=%d, from src=%d to dst=%d\n", length, src, dst)

	// MCOPY pushes nothing onto stack.
	return nil, nil
}

// func enable3860(jt *JumpTable) {
// 	// Overwrite the dynamic gas function for CREATE
// 	jt[CREATE].dynamicGas = gasCreateEIP3860

// 	// Overwrite the dynamic gas function for CREATE2
// 	jt[CREATE2].dynamicGas = gasCreate2EIP3860
// }
