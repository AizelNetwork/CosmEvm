// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/AizelNetwork/evmos/blob/main/LICENSE)
package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m QueryTraceTxRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, msg := range m.Predecessors {
		if msg != nil {
			if err := msg.UnpackInterfaces(unpacker); err != nil {
				return err
			}
		}
	}
	if m.Msg != nil {
		if err := m.Msg.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	// Only attempt to unpack TraceConfig if it implements the interface.
	if m.TraceConfig != nil {
		if uim, ok := interface{}(m.TraceConfig).(codectypes.UnpackInterfacesMessage); ok {
			if err := uim.UnpackInterfaces(unpacker); err != nil {
				return err
			}
		}
		// Otherwise, if TraceConfig doesn't require unpacking, do nothing.
	}
	return nil
}

func (m QueryTraceBlockRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, msg := range m.Txs {
		if err := msg.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

// Failed returns if the contract execution failed in vm errors
func (egr EstimateGasResponse) Failed() bool {
	return len(egr.VmError) > 0
}
