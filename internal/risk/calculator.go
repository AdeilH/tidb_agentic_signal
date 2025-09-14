package risk

import (
	"errors"
	"math"
)

// RiskParams contains risk management parameters
type RiskParams struct {
	AccountBalance  float64 // Total account balance
	RiskPerTrade    float64 // Risk percentage per trade (e.g., 0.02 for 2%)
	MaxPositionSize float64 // Maximum position size percentage (e.g., 0.10 for 10%)
	StopLossPercent float64 // Stop loss percentage (e.g., 0.05 for 5%)
}

// PositionSize calculates the position size based on risk parameters
type PositionSize struct {
	Quantity   float64 // Number of units to trade
	Value      float64 // Total value of position
	StopLoss   float64 // Stop loss price
	RiskAmount float64 // Amount at risk
}

// Calculator provides risk management calculations
type Calculator struct {
	params RiskParams
}

// NewCalculator creates a new risk calculator
func NewCalculator(params RiskParams) (*Calculator, error) {
	if params.AccountBalance <= 0 {
		return nil, errors.New("account balance must be positive")
	}
	if params.RiskPerTrade <= 0 || params.RiskPerTrade > 1 {
		return nil, errors.New("risk per trade must be between 0 and 1")
	}
	if params.MaxPositionSize <= 0 || params.MaxPositionSize > 1 {
		return nil, errors.New("max position size must be between 0 and 1")
	}
	if params.StopLossPercent <= 0 || params.StopLossPercent > 1 {
		return nil, errors.New("stop loss percent must be between 0 and 1")
	}

	return &Calculator{params: params}, nil
}

// CalculatePositionSize calculates optimal position size for a trade
func (c *Calculator) CalculatePositionSize(currentPrice float64, direction string) (*PositionSize, error) {
	if currentPrice <= 0 {
		return nil, errors.New("current price must be positive")
	}

	if direction != "long" && direction != "short" {
		return nil, errors.New("direction must be 'long' or 'short'")
	}

	// Calculate risk amount
	riskAmount := c.params.AccountBalance * c.params.RiskPerTrade

	// Calculate stop loss price
	var stopLoss float64
	if direction == "long" {
		stopLoss = currentPrice * (1 - c.params.StopLossPercent)
	} else {
		stopLoss = currentPrice * (1 + c.params.StopLossPercent)
	}

	// Calculate risk per unit
	riskPerUnit := math.Abs(currentPrice - stopLoss)

	// Calculate quantity based on risk
	quantity := riskAmount / riskPerUnit

	// Calculate position value
	positionValue := quantity * currentPrice

	// Check against maximum position size
	maxPositionValue := c.params.AccountBalance * c.params.MaxPositionSize
	if positionValue > maxPositionValue {
		quantity = maxPositionValue / currentPrice
		positionValue = maxPositionValue
	}

	return &PositionSize{
		Quantity:   math.Floor(quantity*100000) / 100000, // Round to 5 decimal places
		Value:      math.Floor(positionValue*100) / 100,  // Round to 2 decimal places
		StopLoss:   math.Floor(stopLoss*100) / 100,       // Round to 2 decimal places
		RiskAmount: math.Floor(riskAmount*100) / 100,     // Round to 2 decimal places
	}, nil
}

// CalculateRiskReward calculates risk-reward ratio for a trade
func (c *Calculator) CalculateRiskReward(entryPrice, stopLoss, takeProfit float64) (float64, error) {
	if entryPrice <= 0 || stopLoss <= 0 || takeProfit <= 0 {
		return 0, errors.New("all prices must be positive")
	}

	risk := math.Abs(entryPrice - stopLoss)
	reward := math.Abs(takeProfit - entryPrice)

	if risk == 0 {
		return 0, errors.New("risk cannot be zero")
	}

	return reward / risk, nil
}

// ValidateRiskLimits checks if a trade meets risk management criteria
func (c *Calculator) ValidateRiskLimits(positionValue, currentExposure float64) error {
	// Check position size limit
	if positionValue > c.params.AccountBalance*c.params.MaxPositionSize {
		return errors.New("position size exceeds maximum allowed")
	}

	// Check total exposure (including current position)
	totalExposure := currentExposure + positionValue
	maxTotalExposure := c.params.AccountBalance * 0.5 // Max 50% total exposure

	if totalExposure > maxTotalExposure {
		return errors.New("total exposure would exceed 50% of account")
	}

	return nil
}

// GetRiskMetrics returns current risk metrics
func (c *Calculator) GetRiskMetrics() map[string]interface{} {
	return map[string]interface{}{
		"account_balance":    c.params.AccountBalance,
		"risk_per_trade":     c.params.RiskPerTrade * 100,    // Convert to percentage
		"max_position_size":  c.params.MaxPositionSize * 100, // Convert to percentage
		"stop_loss_percent":  c.params.StopLossPercent * 100, // Convert to percentage
		"max_risk_amount":    c.params.AccountBalance * c.params.RiskPerTrade,
		"max_position_value": c.params.AccountBalance * c.params.MaxPositionSize,
	}
}
