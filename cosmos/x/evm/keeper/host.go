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

package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.berachain.dev/polaris/cosmos/config"
	"pkg.berachain.dev/polaris/cosmos/x/evm/plugins/block"
	"pkg.berachain.dev/polaris/cosmos/x/evm/plugins/historical"
	"pkg.berachain.dev/polaris/cosmos/x/evm/plugins/precompile"
	pclog "pkg.berachain.dev/polaris/cosmos/x/evm/plugins/precompile/log"
	"pkg.berachain.dev/polaris/cosmos/x/evm/plugins/state"
	"pkg.berachain.dev/polaris/eth/core"
	ethprecompile "pkg.berachain.dev/polaris/eth/core/precompile"
)

// Compile-time interface assertion.
var _ core.PolarisHostChain = (*Host)(nil)

type Host struct {
	// The various plugins that are are used to implement core.PolarisHostChain.
	bp     block.Plugin
	hp     historical.Plugin
	pp     precompile.Plugin
	sp     state.Plugin
	logger log.Logger

	pcs func() *ethprecompile.Injector
}

// Newhost creates new instances of the plugin host.
func NewHost(
	cfg config.Config,
	storeKey storetypes.StoreKey,
	ak state.AccountKeeper,
	precompiles func() *ethprecompile.Injector,
	qc func() func(height int64, prove bool) (sdk.Context, error),
	logger log.Logger,
) *Host {
	// We setup the host with some Cosmos standard sauce.
	h := &Host{
		bp: block.NewPlugin(
			storeKey, qc,
		),
		pcs:    precompiles,
		pp:     precompile.NewPlugin(),
		sp:     state.NewPlugin(ak, storeKey, qc, nil),
		logger: logger,
	}

	// historical plugin requires block plugin.
	h.hp = historical.NewPlugin(&cfg.Polar.Chain, h.bp, nil, storeKey)
	return h
}

// SetupPrecompiles intializes the precompile contracts.
func (h *Host) SetupPrecompiles() error {
	// Set the query context function for the block and state plugins
	pcs := h.pcs().GetPrecompiles()

	if err := h.pp.RegisterPrecompiles(pcs); err != nil {
		return err
	}

	h.sp.SetPrecompileLogFactory(pclog.NewFactory(pcs))
	return nil
}

// GetBlockPlugin returns the header plugin.
func (h *Host) GetBlockPlugin() core.BlockPlugin {
	return h.bp
}

// GetHistoricalPlugin returns the historical plugin.
func (h *Host) GetHistoricalPlugin() core.HistoricalPlugin {
	return h.hp
}

// GetPrecompilePlugin returns the precompile plugin.
func (h *Host) GetPrecompilePlugin() core.PrecompilePlugin {
	return h.pp
}

// GetStatePlugin returns the state plugin.
func (h *Host) GetStatePlugin() core.StatePlugin {
	return h.sp
}

// GetAllPlugins returns all the plugins.
func (h *Host) GetAllPlugins() []any {
	return []any{h.bp, h.hp, h.pp, h.sp}
}
