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

package core

import (
	"github.com/berachain/stargazer/eth/common"
	"github.com/berachain/stargazer/eth/core/types"
	"github.com/ethereum/go-ethereum/consensus"
)

// Compile-time interface assertion.
var _ ChainContext = (*chainContext)(nil)

// `chainContext` is a wrapper around `StateProcessor` that implements the `ChainContext` interface.
type chainContext struct {
	*StateProcessor
}

// `GetHeader` returns the header for the given hash and height. This is used by the `GetHashFn`.
func (cc *chainContext) GetHeader(_ common.Hash, height uint64) *types.Header {
	if header := cc.StateProcessor.bp.GetStargazerHeaderByNumber(int64(height)); header != nil {
		return header.Header
	}
	return nil
}

// `Engine` returns the consensus engine. For our use case, this never gets called.
func (cc *chainContext) Engine() consensus.Engine {
	return nil
}
