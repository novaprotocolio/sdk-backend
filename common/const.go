package common

import "fmt"

const STATUS_SUCCESSFUL = "successful"
const STATUS_PENDING = "pending"
const STATUS_FAILED = "failed"

func GetMarketOrderbookSnapshotV2Key(marketID string) string {
	return fmt.Sprintf("NOVA_MARKET_ORDERBOOK_SNAPSHOT_V2:%s", marketID)
}

// queue key
const NOVA_WEBSOCKET_MESSAGES_QUEUE_KEY = "NOVA_WEBSOCKET_MESSAGES_QUEUE_KEY"
const NOVA_ENGINE_EVENTS_QUEUE_KEY = "NOVA_ENGINE_EVENTS_QUEUE_KEY"
const NOVA_ENGINE_DEPOSIT_QUEUE_KEY = "NOVA_ENGINE_DEPOSIT_QUEUE_KEY"

// cache key
const NOVA_WATCHER_BLOCK_NUMBER_CACHE_KEY = "NOVA_WATCHER_BLOCK_NUMBER_CACHE_KEY"

// order status
const ORDER_CANCELED = "canceled"
const ORDER_PENDING = "pending"
const ORDER_PARTIAL_FILLED = "partial_filled"
const ORDER_FULL_FILLED = "full_filled"
