package strategy

import "fmt"

func CreateStrategy(name string, parameters map[string]interface{}) (Strategy, error) {
	switch name {
	case "Buy and Hold":
		return NewBuyAndHoldStrategy(), nil
	case "SMA Cross", "SMA Cross (50, 200)":
		fast := 50
		slow := 200
		if fastRaw, ok := parameters["fast"]; ok {
			fast = int(fastRaw.(float64))
		}
		if slowRaw, ok := parameters["slow"]; ok {
			slow = int(slowRaw.(float64))
		}
		return NewSMACrossStrategy(fast, slow), nil
	case "MACD", "MACD (12, 26, 9)":
		return NewMACDStrategy(), nil // Expand with parameters if needed
	case "RSI Mean Reversion":
		return NewMeanReversionStrategy(), nil
	case "Bollinger Bands":
		return NewBollingerBandsStrategy(), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}
