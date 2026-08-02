#!/usr/bin/env sh
set -eu

command -v curl >/dev/null 2>&1 || { echo "curl is required" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }

api_base_url=${API_BASE_URL:-http://localhost:8080}
travel_date=${TRAVEL_DATE:-$(date +%F)}

route_id=$(curl --fail --silent "$api_base_url/api/v1/routes" | jq -er '.items[0].id')
stations=$(curl --fail --silent "$api_base_url/api/v1/routes/$route_id/stations")
origin_id=$(printf '%s' "$stations" | jq -er '.items[0].id')
destination_id=$(printf '%s' "$stations" | jq -er '.items[-1].id')
run_id=$(curl --fail --silent "$api_base_url/api/v1/train-runs?travelDate=$travel_date&routeId=$route_id" | jq -er '.items[0].id')
seat_id=$(curl --fail --silent "$api_base_url/api/v1/train-runs/$run_id/available-seats?originStationId=$origin_id&destinationStationId=$destination_id" | jq -er '.items[0].id')

quote_payload=$(jq -n \
  --arg run "$run_id" --arg seat "$seat_id" --arg origin "$origin_id" --arg destination "$destination_id" \
  '{trainRunId:$run,seatId:$seat,originStationId:$origin,destinationStationId:$destination}')
quote_id=$(curl --fail --silent -X POST "$api_base_url/api/v1/fare-quotes" -H 'Content-Type: application/json' -d "$quote_payload" | jq -er '.id')

booking_payload=$(jq -n \
  --arg quote "$quote_id" --arg run "$run_id" --arg seat "$seat_id" --arg origin "$origin_id" --arg destination "$destination_id" \
  '{fareQuoteId:$quote,trainRunId:$run,seatId:$seat,originStationId:$origin,destinationStationId:$destination,passenger:{fullName:"Integration Test",email:"integration@example.test",phone:"+94770000000"}}')
booking=$(curl --fail --silent -X POST "$api_base_url/api/v1/bookings" \
  -H 'Content-Type: application/json' -H "Idempotency-Key: integration-$(date +%s)-booking" -d "$booking_payload")
booking_id=$(printf '%s' "$booking" | jq -er '.id')
reference=$(printf '%s' "$booking" | jq -er '.reference')

curl --fail --silent "$api_base_url/api/v1/bookings/reference/$reference" | jq -e --arg id "$booking_id" '.id == $id' >/dev/null
curl --fail --silent -X POST "$api_base_url/api/v1/bookings/$booking_id/cancel" | jq -e '.status == "CANCELLED"' >/dev/null

echo "Booking integration flow passed for $reference."
