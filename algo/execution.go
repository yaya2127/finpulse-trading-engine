package algo

import (
	"math"
	"time"

	"github.com/yaya2127/finpulse-trading-engine/engine"
)

type AlgoSlice struct {
	SliceID     uint64  `json:"slice_id"`
	SlicePrice  float64 `json:"slice_price"`
	SliceQty    float64 `json:"slice_qty"`
	ScheduledAt string  `json:"scheduled_at"`
}

// SliceParentOrderVWAP slices a large institutional order using Volume-Weighted Average Price distribution
func SliceParentOrderVWAP(parentQty float64, targetPrice float64, durationMin int, numSlices int) []AlgoSlice {
	slices := make([]AlgoSlice, numSlices)
	baseQty := parentQty / float64(numSlices)

	for i := 0; i < numSlices; i++ {
		// Volume curve modeling (u-shaped volume curve peak at open and close)
		progress := float64(i) / float64(numSlices-1)
		volumeWeight := 1.0 + 0.5*math.Sin(progress*math.Pi)
		sliceQty := baseQty * volumeWeight

		slices[i] = AlgoSlice{
			SliceID:     uint64(i + 1),
			SlicePrice:  targetPrice + (math.Sin(float64(i))*0.25),
			SliceQty:    math.Round(sliceQty*100) / 100,
			ScheduledAt: time.Now().Add(time.Duration(i*durationMin/numSlices) * time.Minute).Format(time.RFC3339),
		}
	}
	return slices
}

// CalculateVWAP computes volume weighted average price for trade history
func CalculateVWAP(trades []engine.TradeExecution) float64 {
	if len(trades) == 0 {
		return 0.0
	}
	var totalValue float64
	var totalVolume float64

	for _, t := range trades {
		totalValue += t.Price * t.Quantity
		totalVolume += t.Quantity
	}

	if totalVolume == 0 {
		return 0.0
	}
	return totalValue / totalVolume
}
