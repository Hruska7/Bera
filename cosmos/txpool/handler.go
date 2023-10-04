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
	"sync/atomic"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/event"

	"pkg.berachain.dev/polaris/eth/core"
	coretypes "pkg.berachain.dev/polaris/eth/core/types"
)

// txChanSize is the size of channel listening to NewTxsEvent. The number is referenced from the
// size of tx pool.
const txChanSize = 4096

// SdkTx is used to generate mocks.
type SdkTx interface {
	sdk.Tx
}

// TxSubProvider.
type TxSubProvider interface {
	SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription
}

// TxSerializer provides an interface to Serialize Geth Transactions to Bytes (via sdk.Tx).
type TxSerializer interface {
	SerializeToSdkTx(signedTx *coretypes.Transaction) (sdk.Tx, error)
	SerializeToBytes(signedTx *coretypes.Transaction) ([]byte, error)
}

// TxBroadcaster provides an interface to broadcast TxBytes to the comet p2p layer.
type TxBroadcaster interface {
	BroadcastTxSync(txBytes []byte) (res *sdk.TxResponse, err error)
}

// Subscription represents a subscription to the txpool.
type Subscription interface {
	event.Subscription
}

// Handler exposes a basic interface to utilize the Handler.
type Handler interface {
	Lifecycle
}

// handler listens for new insertions into the geth txpool and broadcasts them to the CometBFT
// layer for p2p and ABCI.
type handler struct {
	// Cosmos
	logger     log.Logger
	clientCtx  TxBroadcaster
	serializer TxSerializer

	// Ethereum
	txPool  TxSubProvider
	txsCh   chan core.NewTxsEvent
	stopCh  chan struct{}
	txsSub  Subscription
	running atomic.Bool
}

// NewHandler creates a new Handler.
func NewHandler(
	clientCtx TxBroadcaster, txPool TxSubProvider, serializer TxSerializer, logger log.Logger,
) Handler {
	return newHandler(clientCtx, txPool, serializer, logger)
}

// newHandler creates a new handler.
func newHandler(
	clientCtx TxBroadcaster, txPool TxSubProvider, serializer TxSerializer, logger log.Logger,
) *handler {
	h := &handler{
		logger:     logger,
		clientCtx:  clientCtx,
		serializer: serializer,
		txPool:     txPool,
		txsCh:      make(chan core.NewTxsEvent, txChanSize),
		stopCh:     make(chan struct{}),
	}
	return h
}

// Start starts the handler.
func (h *handler) Start() error {
	if h.running.Load() {
		return errors.New("handler already started")
	}
	go h.eventLoop()
	return nil
}

// Stop stops the handler.
func (h *handler) Stop() error {
	if !h.Running() {
		return errors.New("handler already stopped")
	}
	h.stopCh <- struct{}{}
	return nil
}

// start handles the subscription to the txpool and broadcasts transactions.
func (h *handler) eventLoop() {
	// Connect to the subscription.
	h.txsSub = h.txPool.SubscribeNewTxsEvent(h.txsCh)
	h.logger.With("module", "txpool-handler").Info("starting txpool handler")
	h.running.Store(true)

	// Handle events.
	var err error
	for {
		select {
		case <-h.stopCh:
			h.stop(err)
			return
		case err = <-h.txsSub.Err():
			h.stopCh <- struct{}{}
		case event := <-h.txsCh:
			h.broadcastTransactions(event.Txs)
		}
	}
}

// Running returns true if the handler is running.
func (h *handler) Running() bool {
	return h.running.Load()
}

// stop stops the handler.
func (h *handler) stop(err error) {
	// Mark as not running to prevent further events.
	h.running.Store(false)

	// If we are stopping because of an error, log it.
	if err != nil {
		h.logger.Error("txpool handler", "error", err)
	}

	// Triggers txBroadcastLoop to quit.
	h.txsSub.Unsubscribe()

	// Close channels.
	close(h.txsCh)
	close(h.stopCh)
}

// broadcastTransactions will propagate a batch of transactions to the CometBFT mempool.
func (h *handler) broadcastTransactions(txs coretypes.Transactions) {
	h.logger.Debug("broadcasting transactions", "num_txs", len(txs))
	for _, signedEthTx := range txs {
		// Serialize the transaction to Bytes
		txBytes, err := h.serializer.SerializeToBytes(signedEthTx)
		if err != nil {
			h.logger.Error("failed to serialize transaction", "err", err)
			continue
		}

		// Send the transaction to the CometBFT mempool, which will gossip it to peers via
		// CometBFT's p2p layer.
		rsp, err := h.clientCtx.BroadcastTxSync(txBytes)

		// If we see an ABCI response error.
		if rsp != nil && rsp.Code != 0 {
			h.logger.Error("failed to broadcast transaction", "rsp", rsp, "err", err)
		} else if err != nil {
			h.logger.Error("error on transactions broadcast", "err", err)
		}
	}
}
