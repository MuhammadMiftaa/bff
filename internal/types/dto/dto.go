package dto

// ── User data extracted from JWT ──
type UserAuthProvider struct {
	Provider       string `json:"provider"`
	ProviderUserId string `json:"provider_user_id"`
}

type UserData struct {
	ID               string           `json:"id"`
	Email            string           `json:"email"`
	UserAuthProvider UserAuthProvider `json:"user_auth_provider"`
}

// ── Standard API Response ──
type APIResponse struct {
	Status     bool   `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       any    `json:"data,omitempty"`
}

// ── Dashboard HTTP Request DTOs (from frontend) ──

type DateOption struct {
	Date  *string    `json:"date,omitempty"`
	Year  *int       `json:"year,omitempty"`
	Month *int       `json:"month,omitempty"`
	Day   *int       `json:"day,omitempty"`
	Range *DateRange `json:"range,omitempty"`
}

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type GetUserTransactionsRequest struct {
	WalletID   string     `json:"walletID,omitempty"`
	DateOption DateOption `json:"dateOption"`
}

type GetUserBalanceRequest struct {
	WalletID    string     `json:"walletID,omitempty"`
	Aggregation string     `json:"aggregation" validate:"required,oneof=daily weekly monthly"`
	Range       *DateRange `json:"range,omitempty"`
}

type GetUserFinancialSummaryRequest struct {
	WalletID string     `json:"walletID,omitempty"`
	Range    *DateRange `json:"range,omitempty"`
}

// ── Wallet HTTP Request DTOs (from frontend) ──

type CreateWalletRequest struct {
	WalletTypeID string  `json:"wallet_type_id" validate:"required"`
	Name         string  `json:"name" validate:"required"`
	Balance      float64 `json:"balance"`
	Number       string  `json:"number" validate:"required"`
}

type UpdateWalletRequest struct {
	Name         string `json:"name,omitempty"`
	WalletTypeID string `json:"wallet_type_id,omitempty"`
	Number       string `json:"number,omitempty"`
}

// ── Transaction HTTP Request DTOs (from frontend) ──

type CreateTransactionRequest struct {
	WalletID        string   `json:"wallet_id" validate:"required"`
	CategoryID      string   `json:"category_id" validate:"required"`
	Amount          float64  `json:"amount" validate:"required"`
	TransactionDate string   `json:"transaction_date" validate:"required"`
	Description     string   `json:"description,omitempty"`
	Attachments     []string `json:"attachments,omitempty"`
}

type CreateFundTransferRequest struct {
	FromWalletID      string  `json:"from_wallet_id" validate:"required"`
	ToWalletID        string  `json:"to_wallet_id" validate:"required"`
	Amount            float64 `json:"amount" validate:"required"`
	AdminFee          float64 `json:"admin_fee"`
	CashOutCategoryID string  `json:"cash_out_category_id" validate:"required"`
	CashInCategoryID  string  `json:"cash_in_category_id" validate:"required"`
	TransactionDate   string  `json:"transaction_date" validate:"required"`
	Description       string  `json:"description,omitempty"`
}

type UpdateTransactionRequest struct {
	WalletID          string                      `json:"wallet_id,omitempty"`
	CategoryID        string                      `json:"category_id,omitempty"`
	Amount            float64                     `json:"amount,omitempty"`
	TransactionDate   string                      `json:"transaction_date,omitempty"`
	Description       string                      `json:"description,omitempty"`
	AttachmentActions []UpdateAttachmentActionDTO `json:"attachment_actions,omitempty"`
}

type UpdateAttachmentActionDTO struct {
	Status string   `json:"status"` // "add" or "delete"
	Files  []string `json:"files"`
}

type CreateAttachmentRequest struct {
	TransactionID string `json:"transaction_id" validate:"required"`
	Image         string `json:"image" validate:"required"`
	Format        string `json:"format" validate:"required"`
	Size          int64  `json:"size" validate:"required"`
}

// ── Investment HTTP Request DTOs (from frontend) ──

type CreateInvestmentRequest struct {
	Code             string  `json:"code" validate:"required"`
	Quantity         float64 `json:"quantity" validate:"required"`
	Amount           float64 `json:"amount" validate:"required"`
	InitialValuation float64 `json:"initial_valuation"`
	Date             string  `json:"date" validate:"required"`
	Description      string  `json:"description,omitempty"`
}

type SellInvestmentRequest struct {
	AssetCode   string  `json:"asset_code" validate:"required"`
	Quantity    float64 `json:"quantity" validate:"required"`
	Amount      float64 `json:"amount" validate:"required"`
	Date        string  `json:"date" validate:"required"`
	Description string  `json:"description,omitempty"`
}
