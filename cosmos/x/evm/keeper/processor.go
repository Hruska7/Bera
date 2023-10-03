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
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/beacon/engine"

	evmtypes "pkg.berachain.dev/polaris/cosmos/x/evm/types"
	"pkg.berachain.dev/polaris/eth/common"
	coretypes "pkg.berachain.dev/polaris/eth/core/types"
)

func (k *Keeper) ProcessPayloadEnvelope(
	ctx context.Context, msg *evmtypes.WrappedPayloadEnvelope,
) (*evmtypes.WrappedPayloadEnvelopeResponse, error) {
	var envelope = new(engine.ExecutionPayloadEnvelope)
	err := envelope.UnmarshalJSON(msg.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload envelope: %w", err)
	}

	sCtx := sdk.UnwrapSDKContext(ctx)
	gasMeter := sCtx.GasMeter()

	x := new(common.Hash)
	block, err := engine.ExecutableDataToBlock(*envelope.ExecutionPayload, nil, x)
	if err != nil {
		return nil, err
	}

	if err = k.polaris.Blockchain().InsertBlockWithoutSetHead(block); err != nil {
		return nil, err
	}

	// Consume the gas used by the execution of the ethereum block.
	gasMeter.ConsumeGas(block.GasUsed(), "block gas used")

	return &evmtypes.WrappedPayloadEnvelopeResponse{}, nil
}

// ProcessTransaction is called during the DeliverTx processing of the ABCI lifecycle.
func (k *Keeper) ProcessTransaction(ctx context.Context, tx *coretypes.Transaction) (*coretypes.Receipt, error) {
	sCtx := sdk.UnwrapSDKContext(ctx)
	gasMeter := sCtx.GasMeter()
	// We zero-out the gas meter prior to evm execution in order to ensure that the receipt output
	// from the EVM is correct. In the future, we will revisit this to allow gas metering for more
	// complex operations prior to entering the EVM.
	gasMeter.RefundGas(gasMeter.GasConsumed(),
		"reset gas meter prior to ethereum state transition")

	// Process the transaction and return the EVM's execution result.
	receipt, err := k.polaris.SPMiner().ProcessTransaction(ctx, tx)
	if err != nil {
		k.Logger(sCtx).Error("failed to process transaction", "err", err)
		return nil, err
	}

	// Add some safety checks.
	// TODO: we can probably do these once at the end of the block?
	if receipt.GasUsed != gasMeter.GasConsumed() {
		panic(fmt.Sprintf(
			"receipt gas used and ctx gas used differ. receipt: %d, ctx: %d",
			receipt.GasUsed, gasMeter.GasConsumed(),
		))
	} else if receipt.CumulativeGasUsed != sCtx.BlockGasMeter().GasConsumed()+receipt.GasUsed {
		panic(fmt.Sprintf(
			"receipt cumulative gas used and block gas used differ. receipt: %d, ctx: %d",
			receipt.CumulativeGasUsed, sCtx.BlockGasMeter().GasConsumed()+receipt.GasUsed,
		))
	}

	// Log the receipt.
	k.Logger(sCtx).Debug(
		"evm execution completed",
		"tx_hash", receipt.TxHash,
		"gas_consumed", receipt.GasUsed,
		"status", receipt.Status,
	)

	// Return the execution result.
	return receipt, err
}
