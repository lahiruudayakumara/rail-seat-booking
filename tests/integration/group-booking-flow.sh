#!/usr/bin/env sh
set -eu

command -v curl >/dev/null 2>&1 || { echo "curl is required" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }

api_base_url=${API_BASE_URL:-http://localhost:8080}
travel_date=${TRAVEL_DATE:-$(date +%F)}
idempotency_suffix=$(date +%s)-$$

route_id=$(curl --fail --silent "$api_base_url/api/v1/routes" | jq -er '.items[0].id')
stations=$(curl --fail --silent "$api_base_url/api/v1/routes/$route_id/stations?limit=100")
origin_id=$(printf '%s' "$stations" | jq -er '.items[0].id')
destination_id=$(printf '%s' "$stations" | jq -er '.items[-1].id')
run_id=$(curl --fail --silent "$api_base_url/api/v1/train-runs?travelDate=$travel_date&routeId=$route_id" | jq -er '.items[0].id')
seats=$(curl --fail --silent "$api_base_url/api/v1/train-runs/$run_id/available-seats?originStationId=$origin_id&destinationStationId=$destination_id")
seat_one=$(printf '%s' "$seats" | jq -er '.items[0].id')
seat_two=$(printf '%s' "$seats" | jq -er '.items[1].id')

create_member() {
  seat_id=$1
  full_name=$2
  email=$3
  phone=$4
  quote_payload=$(jq -cn --arg run "$run_id" --arg seat "$seat_id" --arg origin "$origin_id" --arg destination "$destination_id" '{trainRunId:$run,seatId:$seat,originStationId:$origin,destinationStationId:$destination}')
  quote=$(curl --fail --silent -X POST "$api_base_url/api/v1/fare-quotes" -H 'Content-Type: application/json' -d "$quote_payload")
  quote_id=$(printf '%s' "$quote" | jq -er '.id')
  hold=$(curl --fail --silent -X POST "$api_base_url/api/v1/booking-holds" -H 'Content-Type: application/json' -d "{\"fareQuoteId\":\"$quote_id\"}")
  jq -cn --arg hold "$(printf '%s' "$hold" | jq -er '.id')" --arg token "$(printf '%s' "$hold" | jq -er '.managementToken')" --arg quote "$quote_id" --arg run "$run_id" --arg seat "$seat_id" --arg origin "$origin_id" --arg destination "$destination_id" --arg name "$full_name" --arg email "$email" --arg phone "$phone" '{holdId:$hold,holdToken:$token,fareQuoteId:$quote,trainRunId:$run,seatId:$seat,originStationId:$origin,destinationStationId:$destination,passenger:{fullName:$name,email:$email,phone:$phone}}'
}

member_one=$(create_member "$seat_one" "Group Integration Lead" "group-lead@example.test" "+94770000001")
member_two=$(create_member "$seat_two" "Group Integration Companion" "group-companion@example.test" "+94770000002")
group_payload=$(jq -cn --argjson first "$member_one" --argjson second "$member_two" '{members:[$first,$second]}')
group=$(curl --fail --silent -X POST "$api_base_url/api/v1/booking-groups" -H 'Content-Type: application/json' -H "Idempotency-Key: group-create-$idempotency_suffix" -d "$group_payload")
group_id=$(printf '%s' "$group" | jq -er '.id')
group_token=$(printf '%s' "$group" | jq -er '.managementToken')

checkout_payload=$(jq -cn --arg id "$group_id" --arg token "$group_token" '{groupId:$id,groupToken:$token}')
checkout=$(curl --fail --silent -X POST "$api_base_url/api/v1/payments/sandbox/groups" -H 'Content-Type: application/json' -H "Idempotency-Key: group-payment-$idempotency_suffix" -d "$checkout_payload")

printf '%s' "$checkout" | jq -e '
  .group.reference | startswith("GR-")
' >/dev/null
printf '%s' "$checkout" | jq -e '
  .group.status == "CONFIRMED" and
  (.group.members | length) == 2 and
  ([.group.members[].booking.status] | all(. == "CONFIRMED")) and
  .payment.status == "PAID" and
  (.tickets | length) == 2 and
  ([.tickets[].status] | all(. == "ACTIVE"))
' >/dev/null

echo "Group booking integration flow passed for $(printf '%s' "$checkout" | jq -r '.group.reference')."
