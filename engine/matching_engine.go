package engine

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type TradeExecution struct {
	BuyOrderID  uint64  `json:"buy_order_id"`
	SellOrderID uint64  `json:"sell_order_id"`
	Price       float64 `json:"price"`
	Quantity    float64 `json:"quantity"`
	ExecutedAt  string  `json:"executed_at"`
}

type OrderBook struct {
	mu     sync.RWMutex
	Symbol string         `json:"symbol"`
	Bids   []OrderMessage `json:"bids"`
	Asks   []OrderMessage `json:"asks"`
	Trades []TradeExecution `json:"trades"`
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make([]OrderMessage, 0),
		Asks:   make([]OrderMessage, 0),
		Trades: make([]TradeExecution, 0),
	}
}

func (ob *OrderBook) ProcessOrder(order OrderMessage) []TradeExecution {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	executedTrades := make([]TradeExecution, 0)

	if order.Side == "BUY" {
		// Match against lowest asks
		for i := 0; i < len(ob.Asks) && order.Quantity > 0; {
			ask := &ob.Asks[i]
			if order.OrderType == "LIMIT" && order.Price < ask.Price {
				break
			}

			tradeQty := order.Quantity
			if ask.Quantity < tradeQty {
				tradeQty = ask.Quantity
			}

			trade := TradeExecution{
				BuyOrderID:  order.ID,
				SellOrderID: ask.ID,
				Price:       ask.Price,
				Quantity:    tradeQty,
				ExecutedAt:  time.Now().Format(time.RFC3339),
			}
			executedTrades = append(executedTrades, trade)
			ob.Trades = append(ob.Trades, trade)

			order.Quantity -= tradeQty
			ask.Quantity -= tradeQty

			if ask.Quantity == 0 {
				ob.Asks = append(ob.Asks[:i], ob.Asks[i+1:]...)
			} else {
				i++
			}
		}

		if order.Quantity > 0 && order.OrderType == "LIMIT" {
			ob.Bids = append(ob.Bids, order)
			sort.Slice(ob.Bids, func(i, j int) bool {
				return ob.Bids[i].Price > ob.Bids[j].Price // Bids descending
			})
		}
	} else if order.Side == "SELL" {
		// Match against highest bids
		for i := 0; i < len(ob.Bids) && order.Quantity > 0; {
			bid := &ob.Bids[i]
			if order.OrderType == "LIMIT" && order.Price > bid.Price {
				break
			}

			tradeQty := order.Quantity
			if bid.Quantity < tradeQty {
				tradeQty = bid.Quantity
			}

			trade := TradeExecution{
				BuyOrderID:  bid.ID,
				SellOrderID: order.ID,
				Price:       bid.Price,
				Quantity:    tradeQty,
				ExecutedAt:  time.Now().Format(time.RFC3339),
			}
			executedTrades = append(executedTrades, trade)
			ob.Trades = append(ob.Trades, trade)

			order.Quantity -= tradeQty
			bid.Quantity -= tradeQty

			if bid.Quantity == 0 {
				ob.Bids = append(ob.Bids[:i], ob.Bids[i+1:]...)
			} else {
				i++
			}
		}

		if order.Quantity > 0 && order.OrderType == "LIMIT" {
			ob.Asks = append(ob.Asks, order)
			sort.Slice(ob.Asks, func(i, j int) bool {
				return ob.Asks[i].Price < ob.Asks[j].Price // Asks ascending
			})
		}
	}

	return executedTrades
}

func (ob *OrderBook) GetBestBidAsk() (bestBid float64, bestAsk float64) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if len(ob.Bids) > 0 {
		bestBid = ob.Bids[0].Price
	}
	if len(ob.Asks) > 0 {
		bestAsk = ob.Asks[0].Price
	}
	return bestBid, bestAsk
}

func (ob *OrderBook) String() string {
	bid, ask := ob.GetBestBidAsk()
	return fmt.Sprintf("[%s] Best Bid: %.2f | Best Ask: %.2f", ob.Symbol, bid, ask)
}
