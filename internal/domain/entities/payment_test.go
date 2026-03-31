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
