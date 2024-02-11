package apperror

//lint:file-ignore ST1005 strings capitalized
var (
	ErrAccountNotFound      = New("Account not found")
	ErrPositionNotFound     = New("Position not found")
	ErrTokenNotFound        = New("Token not found")
	ErrTokenLimitExceeded   = New("Token limit exceeded")
	ErrRequestError         = New("Request error")
	ErrBalanceLimitExceeded = New("Balance limit exceeded")
	ErrInvalidPositionSide  = New("Invalid position side")
	ErrInsufficientFunds    = New("Insufficient funds")
	ErrSetPositionMode      = New("Error set position mode")
	ErrSetLeverage          = New("Error set leverage")
	ErrSetMarginType        = New("Error set margin type")
	ErrOpenOrdersExists     = New("Open orders exists")
	ErrPositionExists       = New("Position exists")
)
