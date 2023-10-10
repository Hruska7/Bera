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
	"math/big"

	"pkg.berachain.dev/polaris/eth/common"
	"pkg.berachain.dev/polaris/eth/core/types"
	"pkg.berachain.dev/polaris/lib/utils"
)

// ChainReader defines methods that are used to read the state and blocks of the chain.
type ChainReader interface {
	ChainBlockReader
	ChainSubscriber
}

// ChainBlockReader defines methods that are used to read information about blocks in the chain.
type ChainBlockReader interface {
	CurrentHeader() *types.Header
	CurrentBlock() *types.Header
	CurrentFinalBlock() *types.Header
	CurrentSafeBlock() *types.Header
	GetBlock(common.Hash, uint64) *types.Block
	GetReceiptsByHash(common.Hash) types.Receipts
	GetBlockByHash(common.Hash) *types.Block
	GetHeaderByNumber(uint64) *types.Header
	GetHeaderByHash(common.Hash) *types.Header
	GetBlockByNumber(uint64) *types.Block
	GetTransactionLookup(common.Hash) *types.TxLookupEntry
	GetTd(common.Hash, uint64) *big.Int
	HasBlock(common.Hash, uint64) bool
	HasBlockAndState(hash common.Hash, number uint64) bool
}

// =========================================================================
// BlockReader
// =========================================================================

// CurrentHeader returns the current header of the blockchain.
func (bc *blockchain) CurrentHeader() *types.Header {
	block, ok := utils.GetAs[*types.Block](bc.currentBlock.Load())
	if block == nil || !ok {
		return nil
	}
	bc.blockNumCache.Add(block.Number().Uint64(), block)
	bc.blockHashCache.Add(block.Hash(), block)
	return block.Header()
}

// CurrentBlock returns the current header of the blockchain.
func (bc *blockchain) CurrentBlock() *types.Header {
	block, ok := utils.GetAs[*types.Block](bc.currentBlock.Load())
	if block == nil || !ok {
		return nil
	}
	bc.blockNumCache.Add(block.Number().Uint64(), block)
	bc.blockHashCache.Add(block.Hash(), block)
	return block.Header()
}

// CurrentSnapBlock is UNUSED in Polaris.
func (bc *blockchain) CurrentSnapBlock() *types.Header {
	return nil
}

// GetHeadersFrom returns a contiguous segment of headers, in rlp-form, going
// backwards from the given number.
func (bc *blockchain) CurrentFinalBlock() *types.Header {
	fb, ok := utils.GetAs[*types.Block](bc.finalizedBlock.Load())
	if fb == nil || !ok {
		return nil
	}
	bc.blockNumCache.Add(fb.Number().Uint64(), fb)
	bc.blockHashCache.Add(fb.Hash(), fb)
	return fb.Header()
}

// CurrentSafeBlock retrieves the current safe block of the canonical
// chain. The block is retrieved from the blockchain's internal cache.
func (bc *blockchain) CurrentSafeBlock() *types.Header {
	// TODO: determine the difference between safe and final in polaris.
	return bc.CurrentFinalBlock()
}

// GetBlock returns a block by its hash or number.
func (bc *blockchain) GetBlock(hash common.Hash, number uint64) *types.Block {
	if block := bc.GetBlockByHash(hash); block != nil {
		return block
	}

	return bc.GetBlockByNumber(number)
}

// GetBlockByHash retrieves a block from the database by hash, caching it if found.
func (bc *blockchain) GetBlockByHash(hash common.Hash) *types.Block {
	// check the block hash cache
	if block, ok := bc.blockHashCache.Get(hash); ok {
		bc.blockNumCache.Add(block.Number().Uint64(), block)
		return block
	}

	// check if historical plugin is supported by host chain
	if bc.hp == nil {
		bc.logger.Debug("historical plugin not supported by host chain")
		return nil
	}

	// check the historical plugin
	block, err := bc.hp.GetBlockByHash(hash)
	if block == nil || err != nil {
		bc.logger.Debug("failed to get block from historical plugin", "block", block, "err", err)
		return nil
	}

	// Cache the found block for next time and return
	bc.blockNumCache.Add(block.Number().Uint64(), block)
	bc.blockHashCache.Add(hash, block)
	return block
}

// GetBlock retrieves a block from the database by hash and number, caching it if found.
func (bc *blockchain) GetBlockByNumber(number uint64) *types.Block {
	// check the block number cache
	if block, ok := bc.blockNumCache.Get(number); ok {
		bc.blockHashCache.Add(block.Hash(), block)
		return block
	}

	var block *types.Block
	if number == 0 {
		// get the genesis block header
		header, err := bc.bp.GetHeaderByNumber(number)
		if header == nil || err != nil {
			return nil
		}
		block = types.NewBlockWithHeader(header)
	} else {
		var err error
		// check if historical plugin is supported by host chain
		if bc.hp == nil {
			bc.logger.Debug("historical plugin not supported by host chain")
			return nil
		}

		// check the historical plugin
		block, err = bc.hp.GetBlockByNumber(number)
		if block == nil || err != nil {
			return nil
		}
	}

	// Cache the found block for next time and return
	bc.blockNumCache.Add(number, block)
	bc.blockHashCache.Add(block.Hash(), block)
	return block
}

// GetReceipts gathers the receipts that were created in the block defined by
// the given hash.
func (bc *blockchain) GetReceiptsByHash(blockHash common.Hash) types.Receipts {
	// check the cache
	if receipts, ok := bc.receiptsCache.Get(blockHash); ok {
		derived, err := bc.deriveReceipts(receipts, blockHash)
		if err != nil {
			bc.logger.Error("failed to derive receipts", "err", err)
			return nil
		}
		return derived
	}

	// check if historical plugin is supported by host chain
	if bc.hp == nil {
		bc.logger.Debug("historical plugin not supported by host chain")
		return nil
	}

	// check the historical plugin
	receipts, err := bc.hp.GetReceiptsByHash(blockHash)
	if receipts == nil || err != nil {
		bc.logger.Debug(
			"failed to get receipts from historical plugin", "receipts", receipts, "err", err)
		return nil
	}

	// cache the found receipts for next time and return
	bc.receiptsCache.Add(blockHash, receipts)
	derived, err := bc.deriveReceipts(receipts, blockHash)
	if err != nil {
		bc.logger.Error("failed to derive receipts", "err", err)
		return nil
	}

	return derived
}

// GetTransaction gets a transaction by hash. It also returns the block hash of the
// block that the transaction was included in, the block number, and the index of the
// transaction in the block. It only retrieves transactions that are included in the chain
// and does not acquire transactions that are in the mempool.
func (bc *blockchain) GetTransactionLookup(
	hash common.Hash,
) *types.TxLookupEntry {
	// check the cache
	if txLookupEntry, ok := bc.txLookupCache.Get(hash); ok {
		return txLookupEntry
	}

	// check if historical plugin is supported by host chain
	if bc.hp == nil {
		bc.logger.Debug("historical plugin not supported by host chain")
		return nil
	}

	// check the historical plugin
	txLookupEntry, err := bc.hp.GetTransactionByHash(hash)
	if err != nil {
		bc.logger.Debug("failed to get transaction by hash", "tx", hash, "err", err)
		return nil
	}

	// cache the found transaction for next time and return
	bc.txLookupCache.Add(hash, txLookupEntry)
	return txLookupEntry
}

// GetHeaderByNumber retrieves a header from the blockchain.
func (bc *blockchain) GetHeaderByNumber(number uint64) *types.Header {
	header, _ := bc.bp.GetHeaderByNumber(number)
	return header
}

// GetHeaderByHash retrieves a block header from the database by hash, caching it if
// found.
func (bc *blockchain) GetHeaderByHash(hash common.Hash) *types.Header {
	header, err := bc.bp.GetHeaderByHash(hash)
	if err != nil && bc.hp != nil {
		// try searching the historical plugin if the block plugin does not have the header
		var block *types.Block
		block, err = bc.hp.GetBlockByHash(hash)
		if err != nil {
			return nil
		}
		header = block.Header()
	}

	return header
}

// GetTd retrieves a block's total difficulty in the canonical chain from the
// database by hash and number, caching it if found.
func (bc *blockchain) GetTd(hash common.Hash, number uint64) *big.Int {
	block := bc.GetBlock(hash, number)
	if block == nil {
		return nil
	}
	return block.Difficulty()
}

// HasBlock returns true if the blockchain contains a block with the given
// hash or number.
func (bc *blockchain) HasBlock(hash common.Hash, number uint64) bool {
	b := bc.GetBlockByNumber(number)
	if b == nil {
		b = bc.GetBlockByHash(hash)
	}
	return b != nil
}

// HasBlockAndState checks if a block and associated state trie is fully present
// in the database or not, caching it if present.
func (bc *blockchain) HasBlockAndState(hash common.Hash, number uint64) bool {
	// Since no state trie, for now, we assume state exists.
	return bc.HasBlock(hash, number)
}

// WriteGenesisBlock writes the genesis block to the database.
func (bc *blockchain) WriteGenesisBlock(block *types.Block) error {
	if err := bc.bp.StoreHeader(block.Header()); err != nil {
		return err
	}
	// todo: this should be write state
	return bc.InsertBlockInternal(block, nil, nil)
}
