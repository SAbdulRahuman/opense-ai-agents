package broker

import (
	"fmt"
	"strings"

	"github.com/seenimoa/openseai/pkg/models"
)

// ════════════════════════════════════════════════════════════════════
// Order Validation
// ════════════════════════════════════════════════════════════════════

// ValidationError represents an order validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult holds the results of order validation.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// IsValid returns true if the order passed all validation checks.
func (v *ValidationResult) IsValid() bool {
	return v.Valid && len(v.Errors) == 0
}

// ErrorString returns a combined error string.
func (v *ValidationResult) ErrorString() string {
	if v.IsValid() {
		return ""
	}
	msgs := make([]string, len(v.Errors))
	for i, e := range v.Errors {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

// ValidateOrder validates an OrderRequest for basic correctness.
func ValidateOrder(req models.OrderRequest) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Ticker is required
	if req.Ticker == "" {
		result.addError("ticker", "ticker is required")
	}

	// Exchange must be NSE, BSE, or NFO
	exchange := strings.ToUpper(req.Exchange)
	if exchange != "NSE" && exchange != "BSE" && exchange != "NFO" {
		result.addError("exchange", fmt.Sprintf("invalid exchange %q, must be NSE, BSE, or NFO", req.Exchange))
	}

	// Side must be BUY or SELL
	if req.Side != models.Buy && req.Side != models.Sell {
		result.addError("side", fmt.Sprintf("invalid order side %q", req.Side))
	}

	// OrderType must be valid
	switch req.OrderType {
	case models.Market, models.Limit, models.SL, models.SLM:
		// valid
	default:
		result.addError("order_type", fmt.Sprintf("invalid order type %q", req.OrderType))
	}

	// Product must be valid
	switch req.Product {
	case models.CNC, models.MIS, models.NRML:
		// valid
	default:
		result.addError("product", fmt.Sprintf("invalid product %q", req.Product))
	}

	// Quantity must be positive
	if req.Quantity <= 0 {
		result.addError("quantity", "quantity must be positive")
	}

	// Price validation based on order type
	if req.OrderType == models.Limit && req.Price <= 0 {
		result.addError("price", "limit orders require a positive price")
	}

	// Trigger price required for SL/SL-M orders
	if (req.OrderType == models.SL || req.OrderType == models.SLM) && req.TriggerPrice <= 0 {
		result.addError("trigger_price", "stop-loss orders require a positive trigger price")
	}

	// Trigger price validation for SL (limit) orders — must also have price
	if req.OrderType == models.SL && req.Price <= 0 {
		result.addError("price", "SL orders require both price and trigger_price")
	}

	// Price sanity — price should not be negative
	if req.Price < 0 {
		result.addError("price", "price cannot be negative")
	}

	// F&O product check — NRML only on NFO exchange
	if req.Product == models.NRML && exchange != "NFO" {
		result.addError("product", "NRML product is only valid on NFO exchange")
	}

	return result
}

// ValidateStopLoss checks that a stop-loss is logically valid.
func ValidateStopLoss(side models.OrderSide, entryPrice, stopLoss float64) error {
	if stopLoss <= 0 {
		return fmt.Errorf("stop_loss must be positive")
	}
	if side == models.Buy && stopLoss >= entryPrice {
		return fmt.Errorf("for BUY orders, stop_loss (%.2f) must be below entry price (%.2f)", stopLoss, entryPrice)
	}
	if side == models.Sell && stopLoss <= entryPrice {
		return fmt.Errorf("for SELL orders, stop_loss (%.2f) must be above entry price (%.2f)", stopLoss, entryPrice)
	}
	return nil
}

// ValidateTarget checks that a target price is logically valid.
func ValidateTarget(side models.OrderSide, entryPrice, target float64) error {
	if target <= 0 {
		return fmt.Errorf("target must be positive")
	}
	if side == models.Buy && target <= entryPrice {
		return fmt.Errorf("for BUY orders, target (%.2f) must be above entry price (%.2f)", target, entryPrice)
	}
	if side == models.Sell && target >= entryPrice {
		return fmt.Errorf("for SELL orders, target (%.2f) must be below entry price (%.2f)", target, entryPrice)
	}
	return nil
}

// ValidateModifyOrder checks that an order modification is valid.
func ValidateModifyOrder(current *models.Order, req models.OrderRequest) error {
	if current == nil {
		return ErrOrderNotFound
	}
	// Can only modify PENDING or OPEN orders
	if current.Status != models.OrderPending && current.Status != models.OrderOpen {
		return fmt.Errorf("%w: order is %s", ErrOrderCantModify, current.Status)
	}
	// Cannot change side
	if req.Side != "" && req.Side != current.Side {
		return fmt.Errorf("cannot change order side from %s to %s", current.Side, req.Side)
	}
	// Quantity must be positive if specified
	if req.Quantity < 0 {
		return fmt.Errorf("modified quantity must be non-negative")
	}
	return nil
}

// addError appends a validation error and marks the result invalid.
func (v *ValidationResult) addError(field, message string) {
	v.Valid = false
	v.Errors = append(v.Errors, ValidationError{Field: field, Message: message})
}
