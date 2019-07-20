package engine

import (
	"context"
	"github.com/novaprotocolio/sdk-backend/common"
	"sync"
)

type Engine struct {
	marketHandlerMap map[string]*MarketHandler

	// Wait for all queue handler exit gracefully
	Wg sync.WaitGroup

	// global ctx, if this ctx is canceled, queue handlers should exit in a short time.
	ctx context.Context

	dbHandler                  *DBHandler
	orderBookSnapshotHandler   *OrderbookSnapshotHandler
	orderBookActivitiesHandler *OrderbookActivitiesHandler

	lock sync.Mutex
}

func NewEngine(ctx context.Context) *Engine {
	engine := &Engine{
		ctx:              ctx,
		marketHandlerMap: make(map[string]*MarketHandler),
		Wg:               sync.WaitGroup{},
	}

	return engine
}

func (e *Engine) RegisterDBHandler(handler DBHandler) {
	e.dbHandler = &handler
}
func (e *Engine) RegisterOrderbookSnapshotHandler(handler OrderbookSnapshotHandler) {
	e.orderBookSnapshotHandler = &handler
}
func (e *Engine) RegisterOrderbookActivitiesHandler(handler OrderbookActivitiesHandler) {
	e.orderBookActivitiesHandler = &handler
}

type DBHandler interface {
	Update(matchResult common.MatchResult) sync.WaitGroup
}
type OrderbookSnapshotHandler interface {
	Update(key string, snapshot *common.SnapshotV2) sync.WaitGroup
}
type OrderbookActivitiesHandler interface {
	Update(webSocketMessages []common.WebSocketMessage) sync.WaitGroup
}

func (e *Engine) HandleNewOrder(order *common.MemoryOrder) (matchResult common.MatchResult, hasMatch bool) {
	e.lock.Lock()
	defer e.lock.Unlock()

	// find or create marketHandler if not exist yet
	if _, exist := e.marketHandlerMap[order.MarketID]; !exist {
		marketHandler, err := NewMarketHandler(e.ctx, order.MarketID)
		if err != nil {
			panic(err)
		}

		e.marketHandlerMap[order.MarketID] = marketHandler
	}

	// feed the handler with this new order
	handler, _ := e.marketHandlerMap[order.MarketID]
	matchResult, hasMatch = handler.handleNewOrder(order)

	e.triggerDBHandlerIfNotNil(matchResult)
	e.triggerOrderbookSnapshotHandlerIfNotNil(handler)
	e.triggerOrderbookActivityHandlerIfNotNil(matchResult.OrderbookActivities)

	return
}

func (e *Engine) ReInsertOrder(order *common.MemoryOrder) (msg *common.WebSocketMessage) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if _, exist := e.marketHandlerMap[order.MarketID]; !exist {
		marketHandler, err := NewMarketHandler(e.ctx, order.MarketID)
		if err != nil {
			panic(err)
		}

		e.marketHandlerMap[order.MarketID] = marketHandler
	}

	handler, _ := e.marketHandlerMap[order.MarketID]
	event := handler.orderbook.InsertOrder(order)

	e.triggerOrderbookSnapshotHandlerIfNotNil(handler)

	if event == nil {
		return
	} else {
		msg := common.OrderbookChangeMessage(handler.market, handler.orderbook.Sequence, event.Side, event.Price, event.Amount)
		return &msg
	}
}

func (e *Engine) HandleCancelOrder(order *common.MemoryOrder) (msg *common.WebSocketMessage, success bool) {
	e.lock.Lock()
	defer e.lock.Unlock()

	handler, _ := e.marketHandlerMap[order.MarketID]

	event := handler.handleCancelOrder(order)
	if event == nil {
		return
	} else {
		e.triggerOrderbookSnapshotHandlerIfNotNil(handler)

		msg := common.OrderbookChangeMessage(handler.market, handler.orderbook.Sequence, event.Side, event.Price, event.Amount)
		return &msg, true
	}
}

func (e *Engine) triggerDBHandlerIfNotNil(matchResult common.MatchResult) {
	if e.dbHandler != nil {
		(*e.dbHandler).Update(matchResult)
	}
}

func (e *Engine) triggerOrderbookSnapshotHandlerIfNotNil(handler *MarketHandler) {
	if e.orderBookSnapshotHandler != nil {
		snapshot := handler.orderbook.SnapshotV2()
		snapshot.Sequence = handler.orderbook.Sequence

		snapshotKey := common.GetMarketOrderbookSnapshotV2Key(handler.market)

		(*e.orderBookSnapshotHandler).Update(snapshotKey, snapshot)
	}
}

func (e *Engine) triggerOrderbookActivityHandlerIfNotNil(msgs []common.WebSocketMessage) {
	if e.orderBookActivitiesHandler != nil {
		(*e.orderBookActivitiesHandler).Update(msgs)
	}
}
