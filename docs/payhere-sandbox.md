# PayHere Sandbox

The application supports PayHere's hosted Sandbox checkout while retaining the built-in payment simulator as the default. The browser never receives the merchant secret. The API signs checkout fields, and only a signature-verified server callback can change a booking from `HELD` to `CONFIRMED` and issue a ticket.

## Prerequisites

1. Create or use a PayHere Sandbox merchant account.
2. Obtain its Merchant ID and Merchant Secret.
3. Expose the local API callback through a public HTTPS tunnel or development ingress. PayHere cannot notify `localhost`.
4. Do not commit `.env`; it is ignored by Git.

## Configuration

Copy `.env.example` to `.env`, then set:

```dotenv
PAYMENT_PROVIDER=payhere
VITE_PAYMENT_PROVIDER=payhere
PAYHERE_SANDBOX=true
PAYHERE_MERCHANT_ID=your-sandbox-merchant-id
PAYHERE_MERCHANT_SECRET=your-sandbox-merchant-secret
PAYHERE_RETURN_URL=http://localhost:3000/payment/return
PAYHERE_CANCEL_URL=http://localhost:3000/payment/cancel
PAYHERE_NOTIFY_URL=https://your-public-api.example/api/v1/webhooks/payhere
```

`PAYMENT_PROVIDER` selects the API integration and `VITE_PAYMENT_PROVIDER` selects the matching browser flow. Keep them aligned. The merchant secret is supplied only to the API container and is never a frontend build argument.

Start everything:

```bash
docker compose up --build
```

With the API running in PayHere mode, the local signed-callback integration check is:

```bash
PAYHERE_MERCHANT_ID=your-sandbox-merchant-id \
PAYHERE_MERCHANT_SECRET=your-sandbox-merchant-secret \
node tests/integration/payhere-webhook-flow.mjs
```

This check creates a held booking, starts a checkout, posts a correctly signed success callback, verifies ticket issuance, repeats the callback to prove idempotency, and confirms a forged signature is rejected. It does not charge a card or replace a hosted-checkout test with PayHere.

Book as a guest or registered passenger using both an email address and phone number. After the seat is held, the application posts a server-signed form to PayHere's hosted Sandbox. PayHere returns the browser to the configured return or cancel page, while the API independently waits for the server callback. The return page polls the protected payment-status endpoint and displays a ticket only after that callback is verified.

Use PayHere's currently documented sandbox test cards and OTP values rather than real card details. Do not treat the browser return URL, query parameters, or client state as proof that money was received.

## Failure and retry behavior

- Reusing an idempotency key returns the same checkout attempt.
- A booking can have only one pending payment attempt.
- Duplicate callbacks are accepted safely and do not issue duplicate tickets.
- Invalid signatures and amount/currency mismatches cannot confirm a booking.
- Failed or cancelled callbacks release the held booking.
- A success callback received after the hold expires is recorded as disputed for manual review.

## Current boundary

PayHere refunds are not enabled. Cancelling a PayHere-paid booking returns `501 PAYHERE_REFUND_NOT_CONFIGURED` instead of falsely reporting a simulator refund. Production adoption must add the PayHere Refund API, operational reconciliation, credential rotation, monitoring, and merchant approval before setting `PAYHERE_SANDBOX=false`.
