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

package txpool

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/core/txpool"

	"pkg.berachain.dev/polaris/cosmos/x/evm/types"
	coretypes "pkg.berachain.dev/polaris/eth/core/types"
	"pkg.berachain.dev/polaris/lib/utils"
)

// AnteHandle implements sdk.AnteHandler.
// It is used to determine whether transactions should be ejected
// from the comet mempool.
func (m *Mempool) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	// We only want to eject transactions from comet on recheck.
	if ctx.ExecMode() == sdk.ExecModeReCheck {
		if wet, ok := utils.GetAs[*types.WrappedEthereumTransaction](msgs[0]); ok {
			if m.shouldEject(wet.Unwrap()) {
				return ctx, errors.New("eject from comet mempool")
			}
		}
	}
	return next(ctx, tx, simulate)
}

// shouldEject returns true if the transaction should be ejected from the CometBFT mempool.
func (m *Mempool) shouldEject(tx *coretypes.Transaction) bool {
	txHash := tx.Hash()
	// Ejection conditions
	// 1. If the transaction has been included in a block.
	// TODO: we should add somemore conditons later.
	return m.txpool.Status(txHash) == txpool.TxStatusIncluded
}
