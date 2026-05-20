package config

var (
	USERNAME_KEY string = "username"
	PASSWORD_KEY string = "password"

	MESSAGE_TYPE_GENERAL             string = "general"
	MESSAGE_TYPE_INVITATION          string = "invitation"
	MESSAGE_TYPE_INVITATION_ACCEPTED string = "invitation_accepted"
	MESSAGE_TYPE_INVITATION_REJECTED string = "invitation_rejected"

	REDIS_PAYMENT_INFOS_KEY string = "redis:paymentInfos:"

	STRIPE_SECRET_KEY         string = "STRIPE_SECRET_KEY"
	SOLANA_RPC_ENDPOINT       string = "SOLANA_RPC_ENDPOINT"
	SOLANA_PROGRAM_ID         string = "SOLANA_PROGRAM_ID"
	SOLANA_ORACLE_PUBKEY      string = "SOLANA_ORACLE_PUBKEY"
	SOLANA_ORACLE_PRIVATE_KEY string = "SOLANA_ORACLE_PRIVATE_KEY"
)
