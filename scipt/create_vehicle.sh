#!/bin/bash

# Stop script on any command error
# set -e

# --- Configuration ---
# !!! REPLACE WITH YOUR ACTUAL API BASE URL !!!
API_BASE_URL="http://localhost:8090/api" # Example: might be http://localhost:8080

# --- Existing User Credentials and UUID ---
# !!! REPLACE WITH THE DETAILS OF AN EXISTING USER !!!
EXISTING_USER_EMAIL="bash.user@example.com"
EXISTING_USER_PASSWORD='Pa$$w0rd'
EXISTING_USER_UUID="128db529-67bd-4367-8aec-7ae04fdd1875" # e.g., "a1b2c3d4-e5f6-7890-1234-567890abcdef"

# --- New Vehicle Details ---
# Generate a somewhat unique identifier for testing vehicle details
UNIQUE_ID=$(date +%s)$RANDOM
CHASSIS_NUMBER="BASHCHASSIS${UNIQUE_ID: -6}"
PROD_YEAR=2024
REGISTRATION="ZG${UNIQUE_ID: -5}"
DISTANCE=5000
VEHICLE_MODEL="BashRunner V8"
VEHICLE_TYPE="Truck"
# ownerUuid will be taken from EXISTING_USER_UUID above

# --- Sanity Check ---
if [ -z "$EXISTING_USER_EMAIL" ] || [ "$EXISTING_USER_EMAIL" == "existing.user@example.com" ]; then
    echo "Error: Please configure EXISTING_USER_EMAIL in the script."
    exit 1
fi
if [ -z "$EXISTING_USER_PASSWORD" ] || [ "$EXISTING_USER_PASSWORD" == "usersActualPassword" ]; then
    echo "Error: Please configure EXISTING_USER_PASSWORD in the script."
    exit 1
fi
if [ -z "$EXISTING_USER_UUID" ] || [ "$EXISTING_USER_UUID" == "put-the-existing-user-uuid-here" ]; then
    echo "Error: Please configure EXISTING_USER_UUID in the script."
    exit 1
fi

echo "--- Starting Vehicle Creation Script (Login First) ---"

# --- Step 1: Login User ---
echo "Attempting to log in user: $EXISTING_USER_EMAIL ..."
LOGIN_PAYLOAD=$(printf '{"email": "%s", "password": "%s"}' "$EXISTING_USER_EMAIL" "$EXISTING_USER_PASSWORD")

# Make the API call, capture HTTP status code and response body separately
HTTP_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -X POST "${API_BASE_URL}/auth/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_PAYLOAD")

# Extract status code and body
HTTP_BODY=$(echo "$HTTP_RESPONSE" | sed '$d')                 # Remove last line (status code)
HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tail -n1 | cut -d: -f2) # Extract status code

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
# Common keys: .accessToken, .access_token, .token
ACCESS_TOKEN=$(echo "$HTTP_BODY" | jq -r '.accessToken') # MODIFY KEY '.accessToken' IF NEEDED

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" == "null" ]; then
    echo "Error: Could not extract access token (tried key '.accessToken') from login response."
    echo "Response Body: $HTTP_BODY"
    exit 1
fi

echo "Login successful."
# echo "Access Token: $ACCESS_TOKEN" # Optional: print full token

# --- Step 2: Create Vehicle ---
echo -e "\nAttempting to create vehicle..."
# Use the pre-configured EXISTING_USER_UUID as ownerUuid
VEHICLE_PAYLOAD=$(printf '{
  "ownerUuid": "%s",
  "chassisNumber": "%s",
  "productionYear": %d,
  "registation": "%s",
  "treveledDistance": %d,
  "vehicleModel": "%s",
  "vehicleType": "%s"
}' "$EXISTING_USER_UUID" "$CHASSIS_NUMBER" "$PROD_YEAR" "$REGISTRATION" "$DISTANCE" "$VEHICLE_MODEL" "$VEHICLE_TYPE")

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
