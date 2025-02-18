package v9

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v8types "github.com/AizelNetwork/CosmEvm/x/evm/migrations/v8/types"
	"github.com/AizelNetwork/CosmEvm/x/evm/types"
)

// MigrateStore migrates the x/evm module state from version 8 to version 9.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	var (
		paramsV8 v8types.V7Params
		params   types.Params
	)

	store := ctx.KVStore(storeKey)

	paramsV8Bz := store.Get(types.KeyPrefixParams)
	cdc.MustUnmarshal(paramsV8Bz, &paramsV8)

	// Copy over fields from v8 to the new params.
	params.AllowUnprotectedTxs = paramsV8.AllowUnprotectedTxs
	params.ActiveStaticPrecompiles = paramsV8.ActiveStaticPrecompiles
	params.EVMChannels = paramsV8.EVMChannels
	params.AccessControl.Call.AccessType = types.AccessType(paramsV8.AccessControl.Call.AccessType)
	params.AccessControl.Create.AccessControlList = paramsV8.AccessControl.Create.AccessControlList
	params.AccessControl.Call.AccessControlList = paramsV8.AccessControl.Call.AccessControlList
	params.AccessControl.Create.AccessType = types.AccessType(paramsV8.AccessControl.Create.AccessType)
	params.ExtraEIPs = paramsV8.ExtraEIPs

	// *** NEW: Enable EIP-5656 by appending "ethereum_5656" if not already present ***
	found := false
	for _, eip := range params.ExtraEIPs {
		if eip == "ethereum_5656" {
			found = true
			break
		}
	}
	if !found {
		params.ExtraEIPs = append(params.ExtraEIPs, "ethereum_5656")
	}

	if err := params.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&params)
	store.Set(types.KeyPrefixParams, bz)
	return nil
}
