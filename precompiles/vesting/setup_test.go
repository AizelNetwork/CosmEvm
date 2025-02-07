// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/AizelNetwork/evmos/blob/main/LICENSE)
package vesting_test

import (
	"testing"

	"github.com/AizelNetwork/evmos/v20/precompiles/vesting"
	"github.com/AizelNetwork/evmos/v20/testutil/integration/aizel/factory"
	"github.com/AizelNetwork/evmos/v20/testutil/integration/aizel/grpc"
	testkeyring "github.com/AizelNetwork/evmos/v20/testutil/integration/aizel/keyring"
	"github.com/AizelNetwork/evmos/v20/testutil/integration/aizel/network"
	"github.com/stretchr/testify/suite"
)

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	bondDenom string

	precompile *vesting.Precompile
}

func TestPrecompileUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest(nKeys int) {
	keyring := testkeyring.New(nKeys)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(nw)
	txFactory := factory.New(nw, grpcHandler)

	stakingParams, err := grpcHandler.GetStakingParams()
	bondDenom := stakingParams.Params.BondDenom

	if err != nil {
		panic(err)
	}

	s.bondDenom = bondDenom
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.network = nw

	if s.precompile, err = vesting.NewPrecompile(
		s.network.App.VestingKeeper,
		s.network.App.AuthzKeeper,
	); err != nil {
		panic(err)
	}
}
