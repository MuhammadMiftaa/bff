package data

// Service field logging constants
const (
	MainService        = "main"
	EnvService         = "env"
	HTTPServerService  = "http_server"
	GRPCClientService  = "grpc_client"
	DashboardService   = "dashboard"
	WalletService      = "wallet"
	TransactionService = "transaction"
	InvestmentService  = "investment"
)

// Message field logging constants
const (
	// env / startup
	LogEnvVarMissing = "env_var_missing"

	// http server
	LogHTTPServerStarted        = "http_server_started"
	LogHTTPServerStartFailed    = "http_server_start_failed"
	LogHTTPServerShutdownFailed = "http_server_shutdown_failed"

	// grpc client
	LogGRPCClientSetupStarted = "grpc_client_setup_started"
	LogGRPCClientSetupSuccess = "grpc_client_setup_success"
	LogGRPCClientSetupFailed  = "grpc_client_setup_failed"
	LogGRPCClientClosed       = "grpc_client_closed"

	// shutdown
	LogShutdownSignalReceived = "shutdown_signal_received"
	LogShutdownCompleted      = "shutdown_completed"

	// dashboard handler
	LogGetFinancialSummaryFailed    = "get_financial_summary_failed"
	LogGetUserBalanceFailed         = "get_user_balance_failed"
	LogGetUserTransactionsFailed    = "get_user_transactions_failed"
	LogGetNetWorthCompositionFailed = "get_net_worth_composition_failed"
	LogGetUserWalletsFailed         = "get_user_wallets_failed"

	// wallet handler
	LogGetWalletsFailed       = "get_wallets_failed"
	LogGetWalletByIDFailed    = "get_wallet_by_id_failed"
	LogCreateWalletFailed     = "create_wallet_failed"
	LogUpdateWalletFailed     = "update_wallet_failed"
	LogDeleteWalletFailed     = "delete_wallet_failed"
	LogGetWalletTypesFailed   = "get_wallet_types_failed"
	LogGetWalletSummaryFailed = "get_wallet_summary_failed"

	// transaction handler
	LogGetTransactionsFailed    = "get_transactions_failed"
	LogGetTransactionByIDFailed = "get_transaction_by_id_failed"
	LogCreateTransactionFailed  = "create_transaction_failed"
	LogCreateFundTransferFailed = "create_fund_transfer_failed"
	LogUpdateTransactionFailed  = "update_transaction_failed"
	LogDeleteTransactionFailed  = "delete_transaction_failed"
	LogGetCategoriesFailed      = "get_categories_failed"
	LogGetAttachmentsFailed     = "get_attachments_failed"
	LogCreateAttachmentFailed   = "create_attachment_failed"
	LogDeleteAttachmentFailed   = "delete_attachment_failed"

	// investment handler
	LogGetInvestmentsFailed       = "get_investments_failed"
	LogGetInvestmentDetailFailed  = "get_investment_detail_failed"
	LogCreateInvestmentFailed     = "create_investment_failed"
	LogSellInvestmentFailed       = "sell_investment_failed"
	LogGetInvestmentSummaryFailed = "get_investment_summary_failed"
	LogGetAssetCodesFailed        = "get_asset_codes_failed"
)
