package payment

import (
	"context"
	"crypto/md5" // #nosec G501 -- PayHere mandates MD5 for its legacy checkout signature protocol.
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	payHereSandboxCheckout = "https://sandbox.payhere.lk/pay/checkout"
	payHereLiveCheckout    = "https://www.payhere.lk/pay/checkout"
)

type PayHereConfig struct {
	MerchantID string
	Secret     string
	Sandbox    bool
	ReturnURL  string
	CancelURL  string
	NotifyURL  string
}

type PayHereProvider struct {
	config PayHereConfig
	client *http.Client
}

type PayHereCustomer struct {
	FullName string
	Email    string
	Phone    string
	Address  string
	City     string
}

type PayHereSession struct {
	ActionURL string            `json:"actionUrl"`
	Fields    map[string]string `json:"fields"`
}

type PayHereNotification struct {
	MerchantID       string
	OrderID          string
	PaymentID        string
	Amount           string
	Currency         string
	StatusCode       string
	Signature        string
	Method           string
	StatusMessage    string
	EventFingerprint string
}

func NewPayHereProvider(config PayHereConfig) *PayHereProvider {
	return &PayHereProvider{config: config, client: &http.Client{Timeout: 5 * time.Second}}
}

func (p *PayHereProvider) Enabled() bool {
	return strings.TrimSpace(p.config.MerchantID) != "" && strings.TrimSpace(p.config.Secret) != ""
}

// ValidateNotifyURL prevents a passenger from entering hosted checkout when
// PayHere cannot deliver the authoritative server callback. A GET is expected
// to return 405 because the webhook only accepts POST; any non-5xx response
// proves that the configured public route reaches an HTTP application.
func (p *PayHereProvider) ValidateNotifyURL(ctx context.Context) error {
	parsed, err := url.Parse(strings.TrimSpace(p.config.NotifyURL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("PayHere notify URL must be a public HTTPS URL")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return err
	}
	response, err := p.client.Do(request)
	if err != nil {
		return fmt.Errorf("PayHere notify URL is unreachable: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("PayHere notify URL returned HTTP %d", response.StatusCode)
	}
	return nil
}

func (p *PayHereProvider) CreateSession(paymentID, bookingID uuid.UUID, customer PayHereCustomer, amountMinor int64, currency string) (PayHereSession, error) {
	if !p.Enabled() {
		return PayHereSession{}, errors.New("PayHere is not configured")
	}
	if customer.Email == "" || customer.Phone == "" {
		return PayHereSession{}, errors.New("PayHere requires both passenger email and phone")
	}
	amount := formatPayHereAmount(amountMinor)
	orderID := paymentID.String()
	firstName, lastName := splitPayHereName(customer.FullName)
	returnURL := addQuery(p.config.ReturnURL, paymentID, bookingID)
	cancelURL := addQuery(p.config.CancelURL, paymentID, bookingID)
	actionURL := payHereLiveCheckout
	if p.config.Sandbox {
		actionURL = payHereSandboxCheckout
	}
	fields := map[string]string{
		"merchant_id": p.config.MerchantID,
		"return_url":  returnURL,
		"cancel_url":  cancelURL,
		"notify_url":  p.config.NotifyURL,
		"order_id":    orderID,
		"items":       "Rail booking " + bookingID.String(),
		"currency":    currency,
		"amount":      amount,
		"first_name":  firstName,
		"last_name":   lastName,
		"email":       customer.Email,
		"phone":       customer.Phone,
		"address":     customer.Address,
		"city":        customer.City,
		"country":     "Sri Lanka",
		"custom_1":    bookingID.String(),
		"custom_2":    paymentID.String(),
		"hash":        p.checkoutSignature(orderID, amount, currency),
	}
	return PayHereSession{ActionURL: actionURL, Fields: fields}, nil
}

func (p *PayHereProvider) VerifyNotification(values url.Values) (PayHereNotification, error) {
	notification := PayHereNotification{
		MerchantID:    values.Get("merchant_id"),
		OrderID:       values.Get("order_id"),
		PaymentID:     values.Get("payment_id"),
		Amount:        values.Get("payhere_amount"),
		Currency:      values.Get("payhere_currency"),
		StatusCode:    values.Get("status_code"),
		Signature:     strings.ToUpper(values.Get("md5sig")),
		Method:        values.Get("method"),
		StatusMessage: values.Get("status_message"),
	}
	if notification.MerchantID != p.config.MerchantID || notification.OrderID == "" || notification.StatusCode == "" || notification.Signature == "" {
		return PayHereNotification{}, errors.New("invalid PayHere notification")
	}
	expected := p.notificationSignature(notification)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(notification.Signature)) != 1 {
		return PayHereNotification{}, errors.New("invalid PayHere notification signature")
	}
	notification.EventFingerprint = md5Upper(strings.Join([]string{notification.OrderID, notification.PaymentID, notification.Amount, notification.Currency, notification.StatusCode, notification.Signature}, "|"))
	return notification, nil
}

func (p *PayHereProvider) checkoutSignature(orderID, amount, currency string) string {
	return md5Upper(p.config.MerchantID + orderID + amount + currency + md5Upper(p.config.Secret))
}

func (p *PayHereProvider) notificationSignature(notification PayHereNotification) string {
	return md5Upper(notification.MerchantID + notification.OrderID + notification.Amount + notification.Currency + notification.StatusCode + md5Upper(p.config.Secret))
}

func md5Upper(value string) string {
	digest := md5.Sum([]byte(value)) // #nosec G401 -- required by PayHere's published signature algorithm.
	return strings.ToUpper(hex.EncodeToString(digest[:]))
}

func formatPayHereAmount(amountMinor int64) string {
	return fmt.Sprintf("%d.%02d", amountMinor/100, amountMinor%100)
}

func splitPayHereName(fullName string) (string, string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "Passenger", "Passenger"
	}
	if len(parts) == 1 {
		return parts[0], parts[0]
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func addQuery(raw string, paymentID, bookingID uuid.UUID) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	query := parsed.Query()
	query.Set("paymentId", paymentID.String())
	query.Set("bookingId", bookingID.String())
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
