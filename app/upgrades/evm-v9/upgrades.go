// Copyright Tharsis Labs Ltd.
// SPDX-License-Identifier: ENCL-1.0
//
// Upgrade handler for the EVM module v9 upgrade.
// In this upgrade we enable EIP–5656 (the MCOPY opcode) by updating the EVM parameters.

package v9

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	evmkeeper "github.com/AizelNetwork/CosmEvm/x/evm/keeper"
	evmtypes "github.com/AizelNetwork/CosmEvm/x/evm/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// UpgradeName is the name of this upgrade.
const UpgradeName = "evm-v9"

// CreateUpgradeHandler returns an upgrade handler for the EVM module upgrade to v9.
// It runs our migration (updating parameters to include the new EIP) and then runs the module migrations.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	ek *evmkeeper.Keeper,
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey, // the exported evm store key from the app
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)
		logger.Info("Starting EVM v9 upgrade migration")

		// Run the EVM store migration: update parameters so that the new EIP is enabled.
		if err := MigrateStore(ctx, storeKey, cdc); err != nil {
			return nil, fmt.Errorf("failed to migrate evm store: %w", err)
		}

		// Run all module migrations.
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateStore migrates the EVM module store from v8 to v9.
// In this migration we update the EVM parameters to include the new EIP (EIP-5656).
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	store := ctx.KVStore(storeKey)

	// Retrieve the current EVM parameters.
	paramsBz := store.Get(evmtypes.KeyPrefixParams)
	if paramsBz == nil {
		return fmt.Errorf("evm parameters not found")
	}
	var params evmtypes.Params
	cdc.MustUnmarshal(paramsBz, &params)

	// Check if the new EIP identifier is already present.
	const newEIP = "ethereum_5656"
	found := false
	for _, eip := range params.ExtraEIPs {
		if eip == newEIP {
			found = true
			break
		}
	}
	if !found {
		params.ExtraEIPs = append(params.ExtraEIPs, newEIP)
		ctx.Logger().Info("EIP-5656 enabled", "eip", newEIP)
	} else {
		ctx.Logger().Info("EIP-5656 already enabled", "eip", newEIP)
	}

	// Validate the updated parameters.
	if err := params.Validate(); err != nil {
		return err
	}

	// Marshal and save the updated parameters.
	newParamsBz := cdc.MustMarshal(&params)
	store.Set(evmtypes.KeyPrefixParams, newParamsBz)

	return nil
}
