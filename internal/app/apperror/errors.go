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
	// ErrInsufficientHold     = New("Insufficient hold")
)
