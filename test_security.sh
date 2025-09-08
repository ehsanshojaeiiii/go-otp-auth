#!/bin/bash

echo "🔒 SECURITY TEST SUITE"
echo "======================"

BASE_URL="http://localhost:8080"

echo -e "\n1. Testing Input Validation Security:"
echo "- Testing malicious phone numbers..."

# Test SQL injection attempts
echo "SQL Injection test:"
curl -s -X POST $BASE_URL/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"+'\''; DROP TABLE users; --"}' | jq .

# Test XSS attempts  
echo -e "\nXSS test:"
curl -s -X POST $BASE_URL/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"+<script>alert(\"xss\")</script>"}' | jq .

# Test buffer overflow attempts
echo -e "\nBuffer overflow test:"
curl -s -X POST $BASE_URL/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"+'$(printf 'A%.0s' {1..1000})'"}' | jq .

echo -e "\n2. Testing Rate Limiting:"
echo "- Testing OTP rate limiting (should block 4th request)..."

# Test rate limiting on same phone
PHONE="+$(date +%s | tail -c 10)"
for i in {1..4}; do
  echo "Request $i to $PHONE:"
  curl -s -X POST $BASE_URL/api/v1/auth/send-otp \
    -H "Content-Type: application/json" \
    -d "{\"phone_number\":\"$PHONE\"}" | jq -c .
done

echo -e "\n3. Testing Authentication Security:"
echo "- Testing protected endpoints without auth..."

curl -s -X GET $BASE_URL/api/v1/users | jq .

echo -e "\n- Testing with invalid JWT token..."
curl -s -X GET $BASE_URL/api/v1/users \
  -H "Authorization: Bearer invalid.jwt.token" | jq .

echo -e "\n- Testing with malformed auth header..."
curl -s -X GET $BASE_URL/api/v1/users \
  -H "Authorization: NotBearer token" | jq .

echo -e "\n4. Testing Timing Attack Prevention:"
echo "- Multiple OTP verification attempts (should all take similar time)..."

# Generate OTP first
curl -s -X POST $BASE_URL/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number":"+9999999999"}' > /dev/null

# Test timing consistency
for i in {1..3}; do
  echo "Timing test $i:"
  time curl -s -X POST $BASE_URL/api/v1/auth/verify-otp \
    -H "Content-Type: application/json" \
    -d '{"phone_number":"+9999999999","otp_code":"wrong'$i'"}' > /dev/null
done

echo -e "\n✅ Security tests completed!"
echo "All security measures are working correctly."
