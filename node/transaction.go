package node

import (
	"fmt"

	"github.com/symphonyprotocol/log"
	p2pNode "github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/scb/block"
	"github.com/symphonyprotocol/value_chain/node/diagram"
)

var tLogger = log.GetLogger("TxSyncMiddleware")

type TransactionMiddleware struct {
	*tcp.BaseMiddleware
}

func (t *TransactionMiddleware) Start(ctx *tcp.P2PContext) {
	t.regHandlers()
}

func (t *TransactionMiddleware) regHandlers() {
	t.HandleRequest(diagram.TX_SEND, func(ctx *tcp.P2PContext) {
		var diag diagram.TransactionSendDiagram
		err := ctx.GetDiagram(&diag)
		if err == nil {
			// 1. check if I already recieved this transaction
			if GetValueChainNode().Chain.HasPendingTransaction(diag.Transaction.ID) {
				tLogger.Warn("Pending pool lready has this tx: %v pending packaged.", diag.Transaction.IDString())
			} else {
				// 1. verify
				if diag.Transaction.Verify() {
					tLogger.Trace("Got new pending tx: %v", diag.Transaction.IDString())
					fromID := diag.GetNodeID()
					// 2. broadcast with my nodeId and same diagId
					diag.NodeID = ctx.LocalNode().GetID()
					ctx.BroadcastToNearbyNodes(diag, 20, func(_p *p2pNode.RemoteNode) bool {
						// will not broadcast back to where the msg is from.
						return _p.GetID() != fromID
					})
					// 3. add to my pending tx
					GetValueChainNode().Chain.SavePendingTx(&diag.Transaction)
					// 4. if I'm mining, restart with new pending tx list.
				} else {
					tLogger.Error("Got a tx: %v but cannot verify it.", diag.Transaction.IDString())
				}
			}
		}
	})

	t.HandleRequest(diagram.TX_REQ, func(ctx *tcp.P2PContext) {
		// 1. check if I have this transaction's content

		// 2. if so, send back the content
	})

	t.HandleRequest(diagram.TX_REQ_RES, func(ctx *tcp.P2PContext) {
		// recieve and save to pending trans list
	})
}

func (b *TransactionMiddleware) DashboardData() interface{} {
	return [][]string{
		[]string{"Current Pending Tx Count", fmt.Sprintf("%v", len(block.FindAllUnpackTransaction()))},
	}
}
func (b *TransactionMiddleware) DashboardType() string               { return "table" }
func (b *TransactionMiddleware) DashboardTitle() string              { return "Transaction Syncing" }
func (b *TransactionMiddleware) DashboardTableHasColumnTitles() bool { return false }
