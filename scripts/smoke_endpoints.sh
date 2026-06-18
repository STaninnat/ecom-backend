#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
API_URL="$BASE_URL/v1"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

ADMIN_EMAIL="admin-$(date +%s)@example.test"
USER_EMAIL="user-$(date +%s)@example.test"
PASSWORD="SmokeTest123!"
CATEGORY_NAME="Smoke Category $(date +%s)"
PRODUCT_NAME="Smoke Product $(date +%s)"
MISSING_UUID="00000000-0000-0000-0000-000000000000"
ADMIN_JAR="$TMP_DIR/admin.cookies"
USER_JAR="$TMP_DIR/user.cookies"
GUEST_JAR="$TMP_DIR/guest.cookies"

require_status() {
	local method="$1"
	local path="$2"
	local expected="$3"
	local body="${4:-}"
	local cookie_header="${5:-}"
	local extra_args=()

	if [[ -n "$body" ]]; then
		extra_args+=(-H "Content-Type: application/json" --data "$body")
	fi
	if [[ -n "$cookie_header" ]]; then
		extra_args+=(-H "Cookie: $cookie_header")
	fi

	local status
	status="$(curl -sS -o "$TMP_DIR/response.json" -w "%{http_code}" -X "$method" "${extra_args[@]}" "$API_URL$path")"
	if [[ "$status" != "$expected" ]]; then
		echo "FAIL $method $path expected $expected got $status"
		cat "$TMP_DIR/response.json"
		exit 1
	fi
	echo "PASS $method $path -> $status"
}

cookie_header() {
	local jar="$1"
	awk '$6 == "access_token" {access=$7} $6 == "refresh_token" {refresh=$7} END {printf "access_token=%s; refresh_token=%s", access, refresh}' "$jar"
}

signup() {
	local email="$1"
	local jar="$2"
	local name="$3"
	local status
	status="$(curl -sS -o "$TMP_DIR/signup.json" -D "$TMP_DIR/signup.headers" -c "$jar" -w "%{http_code}" \
		-H "Content-Type: application/json" \
		--data "{\"name\":\"$name\",\"email\":\"$email\",\"password\":\"$PASSWORD\"}" \
		"$API_URL/auth/signup")"
	if [[ "$status" != "201" ]]; then
		echo "FAIL POST /auth/signup expected 201 got $status"
		cat "$TMP_DIR/signup.json"
		exit 1
	fi
	echo "PASS POST /auth/signup ($email) -> $status"
}

docker compose exec -T postgres psql -U ecom -d ecom_dev -c "select 1" >/dev/null

require_status GET /healthz 200
require_status GET /readiness 200
require_status GET /swagger/index.html 200
require_status GET /errorz 500

require_status POST /auth/signup 400 '{"bad":true}'
signup "$ADMIN_EMAIL" "$ADMIN_JAR" "Smoke Admin $(date +%s)"
signup "$USER_EMAIL" "$USER_JAR" "Smoke User $(date +%s)"
require_status POST /auth/signup 400 "{\"name\":\"Smoke User\",\"email\":\"$USER_EMAIL\",\"password\":\"$PASSWORD\"}"
require_status POST /auth/signin 400 '{"email":"missing@example.test","password":"wrong"}'
require_status GET /auth/google/signin 302
require_status GET /auth/google/callback 400

ADMIN_COOKIE="$(cookie_header "$ADMIN_JAR")"
USER_COOKIE="$(cookie_header "$USER_JAR")"

docker compose exec -T postgres psql -U ecom -d ecom_dev \
	-c "update users set role = 'admin' where email = '$ADMIN_EMAIL';" >/dev/null

require_status GET /users/ 401
require_status GET /users/ 200 "" "$USER_COOKIE"
require_status PUT /users/ 200 "{\"name\":\"Smoke Updated\",\"email\":\"$USER_EMAIL\",\"phone\":\"1234567890\",\"address\":\"Local smoke address\"}" "$USER_COOKIE"

require_status POST /categories/ 401 '{"name":"No Auth"}'
require_status POST /categories/ 403 '{"name":"Not Admin"}' "$USER_COOKIE"
require_status POST /categories/ 201 "{\"name\":\"$CATEGORY_NAME\",\"description\":\"Smoke category\"}" "$ADMIN_COOKIE"
require_status GET /categories/ 200

CATEGORY_ID="$(docker compose exec -T postgres psql -U ecom -d ecom_dev -tA -c "select id from categories where name = '$CATEGORY_NAME' limit 1;")"
if [[ -z "$CATEGORY_ID" ]]; then
	echo "FAIL category id was not created"
	exit 1
fi

require_status POST /products/ 403 '{}' "$USER_COOKIE"
PRODUCT_PAYLOAD="{\"category_id\":\"$CATEGORY_ID\",\"name\":\"$PRODUCT_NAME\",\"description\":\"Smoke product\",\"price\":19.99,\"stock\":8,\"image_url\":\"/static/smoke.jpg\"}"
status="$(curl -sS -o "$TMP_DIR/product.json" -w "%{http_code}" -X POST -H "Content-Type: application/json" -H "Cookie: $ADMIN_COOKIE" --data "$PRODUCT_PAYLOAD" "$API_URL/products/")"
if [[ "$status" != "201" ]]; then
	echo "FAIL POST /products/ expected 201 got $status"
	cat "$TMP_DIR/product.json"
	exit 1
fi
PRODUCT_ID="$(sed -n 's/.*"product_id":"\([^"]*\)".*/\1/p' "$TMP_DIR/product.json")"
echo "PASS POST /products/ -> 201"

require_status GET /products/ 200
require_status GET /products/filter 200 '{}'
require_status GET "/products/$PRODUCT_ID" 401
require_status GET "/products/$PRODUCT_ID" 200 "" "$USER_COOKIE"
require_status GET "/products/$MISSING_UUID" 404 "" "$USER_COOKIE"
require_status PUT /products/ 200 "{\"id\":\"$PRODUCT_ID\",\"category_id\":\"$CATEGORY_ID\",\"name\":\"$PRODUCT_NAME Updated\",\"description\":\"Smoke product updated\",\"price\":21.99,\"stock\":8,\"image_url\":\"/static/smoke.jpg\",\"is_active\":true}" "$ADMIN_COOKIE"

require_status POST /guest-cart/items 400 "{\"product_id\":\"$PRODUCT_ID\",\"quantity\":1}"
require_status POST /guest-cart/items 200 "{\"product_id\":\"$PRODUCT_ID\",\"quantity\":1}" "guest_session_id=guest-smoke"
require_status GET /guest-cart/ 200 "" "guest_session_id=guest-smoke"
require_status PUT /guest-cart/items 200 "{\"product_id\":\"$PRODUCT_ID\",\"quantity\":2}" "guest_session_id=guest-smoke"
require_status DELETE /guest-cart/items 200 "{\"product_id\":\"$PRODUCT_ID\"}" "guest_session_id=guest-smoke"
require_status DELETE /guest-cart/ 200 "" "guest_session_id=guest-smoke"

require_status POST /cart/items 401 "{\"product_id\":\"$PRODUCT_ID\",\"quantity\":1}"
require_status POST /cart/items 200 "{\"product_id\":\"$PRODUCT_ID\",\"quantity\":1}" "$USER_COOKIE"
require_status GET /cart/items 200 "" "$USER_COOKIE"
require_status PUT /cart/items 200 "{\"product_id\":\"$PRODUCT_ID\",\"quantity\":2}" "$USER_COOKIE"
require_status DELETE /cart/items 200 "{\"product_id\":\"$PRODUCT_ID\"}" "$USER_COOKIE"
require_status DELETE /cart/ 200 "" "$USER_COOKIE"

ORDER_PAYLOAD="{\"items\":[{\"product_id\":\"$PRODUCT_ID\",\"quantity\":1,\"price\":21.99}],\"payment_method\":\"card\",\"shipping_address\":\"123 Smoke St\",\"contact_phone\":\"1234567890\"}"
status="$(curl -sS -o "$TMP_DIR/order.json" -w "%{http_code}" -X POST -H "Content-Type: application/json" -H "Cookie: $USER_COOKIE" --data "$ORDER_PAYLOAD" "$API_URL/orders/")"
if [[ "$status" != "201" ]]; then
	echo "FAIL POST /orders/ expected 201 got $status"
	cat "$TMP_DIR/order.json"
	exit 1
fi
ORDER_ID="$(sed -n 's/.*"order_id":"\([^"]*\)".*/\1/p' "$TMP_DIR/order.json")"
echo "PASS POST /orders/ -> 201"
require_status GET /orders/user 200 "" "$USER_COOKIE"
require_status GET "/orders/items/$ORDER_ID" 200 "" "$USER_COOKIE"
require_status GET /orders/ 403 "" "$USER_COOKIE"
require_status GET /orders/ 200 "" "$ADMIN_COOKIE"
require_status PUT "/orders/$ORDER_ID/status" 200 '{"status":"paid"}' "$ADMIN_COOKIE"

require_status POST /payments/intent 401 "{\"order_id\":\"$ORDER_ID\",\"currency\":\"usd\"}"
require_status POST /payments/confirm 400 '{"order_id":""}' "$USER_COOKIE"
require_status GET /payments/history 200 "" "$USER_COOKIE"
require_status GET "/payments/$ORDER_ID" 404 "" "$USER_COOKIE"
require_status POST "/payments/$ORDER_ID/refund" 404 "" "$USER_COOKIE"
require_status GET /payments/admin/pending 200 "" "$ADMIN_COOKIE"
require_status POST /payments/webhook 400 '{}'

REVIEW_PAYLOAD="{\"product_id\":\"$PRODUCT_ID\",\"rating\":5,\"comment\":\"Smoke review\",\"media_urls\":[]}"
status="$(curl -sS -o "$TMP_DIR/review.json" -w "%{http_code}" -X POST -H "Content-Type: application/json" -H "Cookie: $USER_COOKIE" --data "$REVIEW_PAYLOAD" "$API_URL/reviews/")"
if [[ "$status" != "201" ]]; then
	echo "FAIL POST /reviews/ expected 201 got $status"
	cat "$TMP_DIR/review.json"
	exit 1
fi
REVIEW_ID="$(sed -n 's/.*"id":"\([^"]*\)".*/\1/p' "$TMP_DIR/review.json")"
echo "PASS POST /reviews/ -> 201"
require_status GET "/reviews/product/$PRODUCT_ID" 200
require_status GET /reviews/user 200 "" "$USER_COOKIE"
require_status GET "/reviews/$REVIEW_ID" 200
require_status PUT "/reviews/$REVIEW_ID" 200 '{"rating":4,"comment":"Updated smoke review","media_urls":[]}' "$USER_COOKIE"
require_status DELETE "/reviews/$REVIEW_ID" 200 "" "$USER_COOKIE"
require_status GET "/reviews/$MISSING_UUID" 404

require_status POST /admin/user/promote 403 '{"user_id":"nobody"}' "$USER_COOKIE"

echo "Smoke endpoint verification passed against $API_URL"
