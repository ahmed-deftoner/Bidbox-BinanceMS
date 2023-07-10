package utils

import (
	"math/rand"
	"time"
)

var (
	Update_needed = false
)

func SelectRandomElement(strings []string) string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(strings))
	return strings[index]
}

func CalculateLongPosFloatingPnLPercentage(entryPrice float64, markPrice float64, quantity float64) float64 {
	pnl := (markPrice - entryPrice) * quantity
	pnlPercentage := (pnl / (entryPrice * quantity)) * 100
	return pnlPercentage
}

func CalculateShortPosFloatingPnLPercentage(entryPrice float64, markPrice float64, quantity float64) float64 {
	pnl := (entryPrice - markPrice) * quantity
	pnlPercentage := (pnl / (entryPrice * quantity)) * 100
	return pnlPercentage
}
