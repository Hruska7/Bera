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

package mempool

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"

	coretypes "github.com/ethereum/go-ethereum/core/types"

	"pkg.berachain.dev/polaris/eth/common"
)

// iterator is used to iterate over the transactions in the sdk mempool.
type iterator struct {
	txs *coretypes.TransactionsByPriceAndNonce

	// serializer converts eth txs to sdk txs when being iterated over.
	serializer SdkTxSerializer

	empty bool
}

func newIterator(serializer SdkTxSerializer) *iterator {
	return &iterator{
		serializer: serializer,
	}
}

func (i *iterator) reset(
	pendingTxs map[common.Address]coretypes.Transactions,
	pendingBaseFee *big.Int,
	signer coretypes.Signer,
) {
	i.txs = coretypes.NewTransactionsByPriceAndNonce(signer, pendingTxs, pendingBaseFee)
	i.empty = false
}

// Tx implements sdkmempool.Iterator.
func (i *iterator) Tx() sdk.Tx {
	ethTx := i.txs.Peek()
	if ethTx == nil {
		// should never hit this case because the immediately before call to Next() should set
		// empty to true
		return nil
	}

	sdkTx, err := i.serializer.SerializeToSdkTx(ethTx)
	if err != nil {
		// TODO: handle nil tx, could cause downstream panic
		// gtp.logger.Error("eth tx could not be serialized to sdk tx:", err)
		return nil
	}

	return sdkTx
}

// Next implements sdkmempool.Iterator.
func (i *iterator) Next() sdkmempool.Iterator {
	i.txs.Shift()

	if i.txs.Peek() == nil {
		i.empty = true
		return nil
	}

	return i
}
