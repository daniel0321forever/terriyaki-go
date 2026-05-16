package entities

import "testing"

func TestNewStripePaymentInfo(t *testing.T) {
	t.Parallel()

	info := NewStripePaymentInfo(
		"user-1",
		"cus_123",
		"pm_123",
		"visa",
		"4242",
		12,
		2030,
	)

	if info == nil {
		t.Fatalf("expected payment info, got nil")
	}
	if info.UserID != "user-1" {
		t.Fatalf("expected userID user-1, got %q", info.UserID)
	}
	if info.StripeCustomerID != "cus_123" {
		t.Fatalf("expected customerID cus_123, got %q", info.StripeCustomerID)
	}
	if info.StripePaymentMethodID != "pm_123" {
		t.Fatalf("expected paymentMethodID pm_123, got %q", info.StripePaymentMethodID)
	}
	if info.Brand != "visa" {
		t.Fatalf("expected brand visa, got %q", info.Brand)
	}
	if info.Last4 != "4242" {
		t.Fatalf("expected last4 4242, got %q", info.Last4)
	}
	if info.ExpMonth != 12 {
		t.Fatalf("expected expMonth 12, got %d", info.ExpMonth)
	}
	if info.ExpYear != 2030 {
		t.Fatalf("expected expYear 2030, got %d", info.ExpYear)
	}
}

func TestNewPaymentMethodInfo(t *testing.T) {
	t.Parallel()

	info := NewPaymentMethodInfo(
		PaymentProviderStripe,
		"card",
		"user-1",
		"cus_123",
		"pm_123",
		"visa",
		"4242",
		12,
		2030,
	)

	if info == nil {
		t.Fatalf("expected payment method info, got nil")
	}
	if info.ProviderCustomerID != "cus_123" {
		t.Fatalf("expected provider customer ref cus_123, got %q", info.ProviderCustomerID)
	}
	if info.ProviderPaymentMethodID != "pm_123" {
		t.Fatalf("expected provider payment method ref pm_123, got %q", info.ProviderPaymentMethodID)
	}
	if info.Provider != PaymentProviderStripe {
		t.Fatalf("expected stripe provider, got %q", info.Provider)
	}
}

func TestNewSolanaPaymentMethodInfo(t *testing.T) {
	t.Parallel()

	info := NewSolanaPaymentMethodInfo("user-1", "devnet", "wallet_abc", "program_xyz")
	if info == nil {
		t.Fatalf("expected solana payment method info, got nil")
	}
	if info.Network != "devnet" {
		t.Fatalf("expected devnet, got %q", info.Network)
	}
	if info.WalletAddress != "wallet_abc" {
		t.Fatalf("expected wallet_abc, got %q", info.WalletAddress)
	}
}

func TestNewSolanaSettlementInfo(t *testing.T) {
	t.Parallel()

	info := NewSolanaSettlementInfo("user-1", "devnet", "sig_123", "contract_abc", SettlementStatusAuthorized, 0)
	if info == nil {
		t.Fatalf("expected solana settlement info, got nil")
	}
	if info.TransactionSignature != "sig_123" {
		t.Fatalf("expected sig_123, got %q", info.TransactionSignature)
	}
	if info.Status != SettlementStatusAuthorized {
		t.Fatalf("expected authorized status, got %q", info.Status)
	}
}

func TestPaymentProviderConstants(t *testing.T) {
	t.Parallel()

	if PaymentProviderStripe != "stripe" {
		t.Fatalf("unexpected stripe provider constant")
	}
	if PaymentProviderSolana != "solana" {
		t.Fatalf("unexpected solana provider constant")
	}
}
