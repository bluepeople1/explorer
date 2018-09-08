package main

import (
	"github.com/iost-official/explorer/backend/task/cron"
	"sync"
)

var ws = new(sync.WaitGroup)

func main()  {
	ws.Add(5)
	go cron.UpdateBlocks(ws)
	go cron.ProcessFailedSyncBlocks(ws)
	go cron.UpdateTxns(ws)
	go cron.UpdateBlockPay(ws)
	go cron.UpdateAccounts(ws)
	ws.Wait()
}
