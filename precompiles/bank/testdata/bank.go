// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/AizelNetwork/evmos/blob/main/LICENSE)

package testdata

import (
	contractutils "github.com/AizelNetwork/CosmEvm/contracts/utils"
	evmtypes "github.com/AizelNetwork/CosmEvm/x/evm/types"
)

func LoadBankCallerContract() (evmtypes.CompiledContract, error) {
	return contractutils.LoadContractFromJSONFile("BankCaller.json")
}
