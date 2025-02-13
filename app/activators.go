// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/AizelNetwork/evmos/blob/main/LICENSE)
package app

import (
	"github.com/AizelNetwork/evmos/v20/app/eips"
	"github.com/AizelNetwork/evmos/v20/x/evm/core/vm"
)

// aizelActivators defines a map of opcode modifiers associated
// with a key defining the corresponding EIP.
var aizelActivators = map[string]func(*vm.JumpTable){
	"aizel_0": eips.Enable0000,
	"aizel_1": eips.Enable0001,
	"aizel_2": eips.Enable0002,
}
