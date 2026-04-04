package cache

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// ── TTL Constants ──

const (
	TTLShort  = 2 * time.Minute  // volatile data: wallet list, summary
	TTLMedium = 5 * time.Minute  // moderately stable: transaction lists, detail
	TTLLong   = 10 * time.Minute // aggregation: dashboard financial summary, balance
	TTLStatic = 1 * time.Hour    // reference data: wallet types, categories
	TTLAsset  = 6 * time.Hour    // very stable: asset codes
)

// ── Dashboard Keys ──

func DashboardFinancialSummary(userID, walletID, rangeStart, rangeEnd string) string {
	return fmt.Sprintf("dashboard:%s:financial-summary:%s:%s:%s", userID, walletID, rangeStart, rangeEnd)
}

func DashboardBalance(userID, walletID, aggregation, rangeStart, rangeEnd string) string {
	return fmt.Sprintf("dashboard:%s:balance:%s:%s:%s:%s", userID, walletID, aggregation, rangeStart, rangeEnd)
}

func DashboardTransactions(userID, walletID, dateOptionHash string) string {
	return fmt.Sprintf("dashboard:%s:transactions:%s:%s", userID, walletID, dateOptionHash)
}

func DashboardCategoryTransactions(userID, walletID, categoryID, dateOptionHash string) string {
	return fmt.Sprintf("dashboard:%s:category-transactions:%s:%s:%s", userID, walletID, categoryID, dateOptionHash)
}

func DashboardNetWorth(userID string) string {
	return fmt.Sprintf("dashboard:%s:net-worth", userID)
}

func DashboardWallets(userID string) string {
	return fmt.Sprintf("dashboard:%s:wallets", userID)
}

// ── Wallet Keys ──

func WalletList(userID string) string {
	return fmt.Sprintf("wallet:%s:list", userID)
}

func WalletSummary(userID string) string {
	return fmt.Sprintf("wallet:%s:summary", userID)
}

func WalletTypes() string {
	return "wallet:types"
}

func WalletByID(walletID string) string {
	return fmt.Sprintf("wallet:%s", walletID)
}

// ── Transaction Keys ──

func TransactionList(userID, paramsHash string) string {
	return fmt.Sprintf("tx:%s:list:%s", userID, paramsHash)
}

func TransactionCategories(categoryType string) string {
	return fmt.Sprintf("tx:categories:%s", categoryType)
}

func TransactionByID(txID string) string {
	return fmt.Sprintf("tx:%s:detail", txID)
}

// ── Investment Keys ──

func InvestmentList(userID, paramsHash string) string {
	return fmt.Sprintf("inv:%s:list:%s", userID, paramsHash)
}

func InvestmentSummary(userID string) string {
	return fmt.Sprintf("inv:%s:summary", userID)
}

func InvestmentAssetCodes() string {
	return "inv:asset-codes"
}

func InvestmentDetail(invID string) string {
	return fmt.Sprintf("inv:%s:detail", invID)
}

// ── Profile Keys ──

func ProfileData(userID string) string {
	return fmt.Sprintf("profile:%s:data", userID)
}

func ProfileAllPattern(userID string) string {
	return fmt.Sprintf("profile:%s:*", userID)
}

// ── Invalidation Patterns ──

func DashboardAllPattern(userID string) string {
	return fmt.Sprintf("dashboard:%s:*", userID)
}

func WalletAllPattern(userID string) string {
	return fmt.Sprintf("wallet:%s:*", userID)
}

func TransactionListPattern(userID string) string {
	return fmt.Sprintf("tx:%s:list:*", userID)
}

func InvestmentAllPattern(userID string) string {
	return fmt.Sprintf("inv:%s:*", userID)
}

// ── Budget Keys ──

func BudgetList(userID, period string) string {
	return fmt.Sprintf("budget:%s:list:%s", userID, period)
}

func BudgetAllPattern(userID string) string {
	return fmt.Sprintf("budget:%s:*", userID)
}

// ── Helper ──

// HashParams produces a short deterministic hash from an arbitrary string (e.g. query params).
func HashParams(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h[:8]) // 16 hex chars
}
