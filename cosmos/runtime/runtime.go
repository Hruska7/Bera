// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2023, Berachain Foundation. All rights reserved.
// Use of this software is govered by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package runtime

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/node"

	libtx "pkg.berachain.dev/polaris/cosmos/lib/tx"
	"pkg.berachain.dev/polaris/cosmos/miner"
	"pkg.berachain.dev/polaris/cosmos/txpool"
	evmkeeper "pkg.berachain.dev/polaris/cosmos/x/evm/keeper"
	enginep "pkg.berachain.dev/polaris/cosmos/x/evm/plugins/engine"
	evmtypes "pkg.berachain.dev/polaris/cosmos/x/evm/types"
	coretypes "pkg.berachain.dev/polaris/eth/core/types"
	"pkg.berachain.dev/polaris/eth/polar"
)

type Polaris struct {
	*polar.Polaris

	// polaris keepers
	EVMKeeper *evmkeeper.Keeper

	// polaris componets
	mm *miner.Miner
	mp *txpool.Mempool
}

func (p *Polaris) Setup(bApp *baseapp.BaseApp) error {
	// SetupPrecompiles is used to setup the precompile contracts post depinject.
	if err := p.EVMKeeper.SetupPrecompiles(); err != nil {
		return err
	}

	// Init is used to setup the polaris struct.
	if err := p.Polaris.Init(); err != nil {
		return err
	}

	// Setup TxPool Wrapper
	p.mp = txpool.New(p.TxPool())
	bApp.SetMempool(p.mp)

	p.mm = miner.New(p.Miner())
	bApp.SetPrepareProposal(p.mm.PrepareProposal)

	// TODO: deprecate this
	p.EVMKeeper.SetBlockchain(p.Blockchain())

	return nil
}

func (p *Polaris) Init(clientCtx client.Context, logger log.Logger) error {
	// Initialize services.
	p.mm.Init(libtx.NewSerializer[*engine.ExecutionPayloadEnvelope](
		clientCtx.TxConfig, evmtypes.WrapPayload))

	p.mp.Init(logger, clientCtx, libtx.NewSerializer[*coretypes.Transaction](
		clientCtx.TxConfig, evmtypes.WrapTx))

	// Register services with Polaris.
	p.RegisterServices(clientCtx, []node.Lifecycle{
		p.mp,
	})
	return nil
}

// Register Services allows for the application to register lifecycles with the evm
// networking stack.
func (p *Polaris) RegisterServices(clientContext client.Context, lcs []node.Lifecycle) {
	// TODO: probably get rid of engine plugin or something and handle rpc methods better.
	p.EVMKeeper.Host.GetEnginePlugin().(enginep.Plugin).Start(clientContext)

	// Register the services with polaris.
	for _, lc := range lcs {
		p.RegisterService(lc)
	}

	// Start the services.
	if err := p.StartServices(); err != nil {
		panic(err)
	}
}

func (p *Polaris) LoadLastState(cms storetypes.CommitMultiStore, appHeight uint64) error {
	cmsCtx := sdk.Context{}.
		WithMultiStore(cms).
		WithGasMeter(storetypes.NewInfiniteGasMeter()).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter()).WithEventManager(sdk.NewEventManager())
	return p.Blockchain().LoadLastState(cmsCtx, appHeight)
}
