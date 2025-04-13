#!/bin/bash

# Stop script on any command error
# set -e

# --- Configuration ---
# !!! REPLACE WITH YOUR ACTUAL API BASE URL !!!
API_BASE_URL="http://localhost:8090/api" # Example: might be http://localhost:8080

# --- New User Details ---
# Generate a somewhat unique identifier for testing
UNIQUE_ID=$(date +%s)$RANDOM
FIRST_NAME="BashTest"
LAST_NAME="User${UNIQUE_ID: -5}"
EMAIL="bash.user.${UNIQUE_ID}@example.com"
PASSWORD="aBashSecurePassword123!" # Use a strong password
OIB="987654321${UNIQUE_ID: -2}"    # Example OIB, ensure format is valid
BIRTH_DATE="1991-02-20"            # Ensure correct ISO 8601 format if needed
RESIDENCE="Bashville, Script Street 101"
ROLE="osoba" # Assuming 'user' is a valid role

# --- New Vehicle Details ---
CHASSIS_NUMBER="BASHCHASSIS${UNIQUE_ID: -6}"
PROD_YEAR=2024
REGISTRATION="ZG${UNIQUE_ID: -5}"
DISTANCE=5000
VEHICLE_MODEL="BashRunner V8"
VEHICLE_TYPE="Truck"
# ownerUuid will be added dynamically

echo "--- Starting Vehicle Creation Script ---"

# --- Step 1: Create User ---
echo "Attempting to create user..."
# Construct JSON payload using printf for safety with quotes
USER_PAYLOAD=$(printf '{
  "firstName": "%s",
  "lastName": "%s",
  "email": "%s",
  "password": "%s",
  "oib": "%s",
  "birthDate": "%s",
  "residence": "%s",
  "role": "%s"
}' "$FIRST_NAME" "$LAST_NAME" "$EMAIL" "$PASSWORD" "$OIB" "$BIRTH_DATE" "$RESIDENCE" "$ROLE")

# Make the API call, capture HTTP status code and response body separately
HTTP_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -X POST "${API_BASE_URL}/user/" \
    -H "Content-Type: application/json" \
    -d "$USER_PAYLOAD")

# Extract status code and body
HTTP_BODY=$(echo "$HTTP_RESPONSE" | sed '$d')                 # Remove last line (status code)
HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tail -n1 | cut -d: -f2) # Extract status code

echo "Create User Status Code: $HTTP_STATUS"
echo "Create User Response Body: $HTTP_BODY"

# Check if user creation was successful (expecting 201 Created)
if [ "$HTTP_STATUS" -ne 201 ]; then
    echo "Error: Failed to create user. Status: $HTTP_STATUS"
    exit 1
fi

# Extract UUID using jq (-r removes quotes)
USER_UUID=$(echo "$HTTP_BODY" | jq -r '.Uuid')

if [ -z "$USER_UUID" ] || [ "$USER_UUID" == "null" ]; then
    echo "Error: Could not extract user UUID from response."
    exit 1
fi

echo "User created successfully. UUID: $USER_UUID"

# --- Step 2: Login User ---
echo -e "\nAttempting to log in user..."
LOGIN_PAYLOAD=$(printf '{"email": "%s", "password": "%s"}' "$EMAIL" "$PASSWORD")

HTTP_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -X POST "${API_BASE_URL}/auth/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_PAYLOAD")

# Extract status code and body
HTTP_BODY=$(echo "$HTTP_RESPONSE" | sed '$d')
HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tail -n1 | cut -d: -f2)

echo "Login Status Code: $HTTP_STATUS"
# echo "Login Response Body: $HTTP_BODY" # Optional: can be verbose

# Check if login was successful (expecting 200 OK)
if [ "$HTTP_STATUS" -ne 200 ]; then
    echo "Error: Failed to log in user. Status: $HTTP_STATUS"
    echo "Response Body: $HTTP_BODY"
    exit 1
fi

# --- IMPORTANT ---
# Extract Access Token using jq. Adjust '.accessToken' if your API uses a different key!
ACCESS_TOKEN=$(echo "$HTTP_BODY" | jq -r '.accessToken') # MODIFY KEY '.accessToken' IF NEEDED

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" == "null" ]; then
    echo "Error: Could not extract access token from login response."
    echo "Response Body: $HTTP_BODY"
    exit 1
fi

echo "Login successful."
# echo "Access Token: $ACCESS_TOKEN" # Optional: print full token

# --- Step 3: Create Vehicle ---
echo -e "\nAttempting to create vehicle..."
VEHICLE_PAYLOAD=$(printf '{
  "ownerUuid": "%s",
  "chassisNumber": "%s",
  "productionYear": %d,
  "registation": "%s",
  "treveledDistance": %d,
  "vehicleModel": "%s",
  "vehicleType": "%s"
}' "$USER_UUID" "$CHASSIS_NUMBER" "$PROD_YEAR" "$REGISTRATION" "$DISTANCE" "$VEHICLE_MODEL" "$VEHICLE_TYPE")

# Note the added -H "Authorization: Bearer $ACCESS_TOKEN"
HTTP_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -X POST "${API_BASE_URL}/vehicle/" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "$VEHICLE_PAYLOAD")

# Extract status code and body
HTTP_BODY=$(echo "$HTTP_RESPONSE" | sed '$d')
HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tail -n1 | cut -d: -f2)

echo "Create Vehicle Status Code: $HTTP_STATUS"
echo "Create Vehicle Response Body: $HTTP_BODY"

# Check if vehicle creation was successful (expecting 201 Created)
if [ "$HTTP_STATUS" -ne 201 ]; then
    echo "Error: Failed to create vehicle. Status: $HTTP_STATUS"
    exit 1
fi

echo "Vehicle created successfully!"

echo -e "\n--- Script Finished ---"
exit 0
