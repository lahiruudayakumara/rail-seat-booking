package payment

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
)

func testPayHereProvider() *PayHereProvider {
	return NewPayHereProvider(PayHereConfig{MerchantID: "1210000", Secret: "sandbox-merchant-secret", Sandbox: true, ReturnURL: "http://localhost:3000/payment/return", CancelURL: "http://localhost:3000/payment/cancel", NotifyURL: "https://example.test/webhooks/payhere"})
}

func TestPayHereSessionUsesSandboxAndNeverExposesSecret(t *testing.T) {
	provider := testPayHereProvider()
	paymentID, bookingID := uuid.New(), uuid.New()
	session, err := provider.CreateSession(paymentID, bookingID, PayHereCustomer{FullName: "Anura Perera", Email: "anura@example.test", Phone: "+94770000000", Address: "1 Main Street", City: "Colombo"}, 69000, "LKR")
	if err != nil {
		t.Fatal(err)
	}
	if session.ActionURL != payHereSandboxCheckout || session.Fields["amount"] != "690.00" || session.Fields["order_id"] != paymentID.String() {
		t.Fatalf("unexpected session: %#v", session)
	}
	for key, value := range session.Fields {
		if value == provider.config.Secret {
			t.Fatalf("merchant secret exposed in field %s", key)
		}
	}
}

func TestPayHereNotificationVerification(t *testing.T) {
	provider := testPayHereProvider()
	values := url.Values{"merchant_id": {"1210000"}, "order_id": {uuid.NewString()}, "payment_id": {"320027150501"}, "payhere_amount": {"690.00"}, "payhere_currency": {"LKR"}, "status_code": {"2"}}
	notification := PayHereNotification{MerchantID: values.Get("merchant_id"), OrderID: values.Get("order_id"), Amount: values.Get("payhere_amount"), Currency: values.Get("payhere_currency"), StatusCode: values.Get("status_code")}
	values.Set("md5sig", provider.notificationSignature(notification))
	verified, err := provider.VerifyNotification(values)
	if err != nil || verified.EventFingerprint == "" {
		t.Fatalf("valid notification rejected: %#v %v", verified, err)
	}
	values.Set("payhere_amount", "1.00")
	if _, err = provider.VerifyNotification(values); err == nil {
		t.Fatal("tampered notification accepted")
	}
}
