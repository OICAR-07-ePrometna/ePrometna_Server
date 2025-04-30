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
EXISTING_USER_UUID="9fea1366-82fe-4775-ba1e-772b77105bb0" # e.g., "a1b2c3d4-e5f6-7890-1234-567890abcdef"

# --- New Vehicle Details ---
# Generate a somewhat unique identifier for testing vehicle details
UNIQUE_ID=$(date +%s)$RANDOM
YEAR=$(date +%Y)
MONTH=$(date +%m)
DAY=$(date +%d)

# Fields already present in the original script
CHASSIS_NUMBER="BASHVIN${UNIQUE_ID: -8}"
PROD_YEAR=$YEAR
REGISTRATION="ZG${UNIQUE_ID: -5}SH"
DISTANCE=5000
VEHICLE_MODEL="BashRunner V8 ${UNIQUE_ID: -3}" # Mapped to "model" key
# ownerUuid will be taken from EXISTING_USER_UUID above

# --> ADDED: Variables for all fields from the Go struct <--
VEHICLE_CATEGORY="M1" # J - Kategorija vozila
MARK="BashMake"       # D1 - Marka
# MODEL uses VEHICLE_MODEL variable above # (14) - Model
HOMOLOGATION_TYPE="BHT${UNIQUE_ID: -4}"           # D2 - Homologacijski tip
TRADE_NAME="Bash Trade ${UNIQUE_ID: -3}"          # D3 - Trgovački naziv
BODY_SHAPE="SUV"                                  # (2) - Oblik karoserije
VEHICLE_USE="Personal"                            # (3) - Namjena vozila
DATE_FIRST_REGISTRATION="$YEAR-$MONTH-$DAY"       # B - Datum prve registracije
FIRST_REGISTRATION_IN_CROATIA="$YEAR-$MONTH-$DAY" # (4) - Prva registracija u Hrvatskoj
TECH_MAX_MASS="2500"                              # F1 - Tehnički dopuštena najveća masa
PERM_MAX_MASS="2400"                              # F2 - Dopuštena najveća masa
UNLADEN_MASS="1800"                               # G - Masa praznog vozila
PERM_PAYLOAD="600"                                # (5) - Dopuštena nosivost
TYPE_APPROVAL_NUMBER="BTA${UNIQUE_ID: -5}"        # K - Broj homologacije
ENGINE_CAPACITY="4999"                            # P1 - Obujam motora
ENGINE_POWER="300"                                # P2 - Snaga motora
FUEL_SOURCE="Petrol"                              # P3 - Gorivo ili izvor energije
RATED_ENGINE_SPEED="6000"                         # P4 - Nazivni broj okretaja motora
NUM_SEATS="5"                                     # S1 - Broj sjedala
COLOUR="Black"                                    # R - Boja vozila
LENGTH="4800"                                     # (6) - Dužina
WIDTH="1900"                                      # (7) - Širina
HEIGHT="1750"                                     # (8) - Visina
MAX_NET_POWER="295"                               # T - Najveća neto snaga
NUM_AXLES="2"                                     # L - Broj osovina
NUM_DRIVEN_AXLES="2"                              # (9) - Broj pogonskih osovina
MB_FIELD="MB Data ${UNIQUE_ID: -2}"               # (13) - MB (unknown meaning, placeholder)
STATIONARY_NOISE="85"                             # U1 - Razina buke u stacionarnom stanju
NOISE_TEST_SPEED="3500"                           # U2 - Broj okretaja motora pri ispitivanju buke
CO2_EMISSIONS="210"                               # V7 - Emisija CO2
EC_CATEGORY="EURO 6"                              # V9 - EC kategorija
TIRE_SIZE="235/55 R19"                            # (11) - Dimenzije guma
UNIQUE_MODEL_CODE="UMC${UNIQUE_ID: -6}"           # (12) - Jedinstvena oznaka modela
ADDITIONAL_TIRE_SIZES="235/60 R18"                # (15) - Dodatne dimenzije guma
VEHICLE_TYPE="Passenger Car SUV"                  # (16) - Tip vozila (also present in original script)
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
# --> UPDATED PAYLOAD with all fields <--
# Using multiple lines inside printf format string for readability

VEHICLE_PAYLOAD=$(
    printf '{
  "ownerUuid": "%s",
  "registration": "%s",
  "traveledDistance": %d,
  "summary": {
    "additionalTireSizes": "%s",
    "bodyShape": "%s",
    "chassisNumber": "%s",
    "co2Emissions": "%s",
    "colourOfVehicle": "%s",
    "dateFirstRegistration": "%s",
    "ecCategory": "%s",
    "engineCapacity": "%s",
    "enginePower": "%s",
    "engineSpeedForStationaryNoiseTest": "%s",
    "firstRegistrationInCroatia": "%s",
    "fuelOrPowerSource": "%s",
    "height": "%s",
    "homologationType": "%s",
    "length": "%s",
    "mark": "%s",
    "maximumNetPower": "%s",
    "mb": "%s",
    "model": "%s",
    "numberOfAxles": "%s",
    "numberOfDrivenAxles": "%s",
    "numberOfSeats": "%s",
    "permissibleMaximumLadenMass": "%s",
    "permissiblePayload": "%s",
    "ratedEngineSpeed": "%s",
    "stationaryNoiseLevel": "%s",
    "technicallyPermissibleMaximumLadenMass": "%s",
    "tireSize": "%s",
    "tradeName": "%s",
    "typeApprovalNumber": "%s",
    "uniqueModelCode": "%s",
    "unladenMass": "%s",
    "vehicleCategory": "%s",
    "vehicleType": "%s",
    "vehicleUse": "%s",
    "width": "%s"
  }
}' \
        "$EXISTING_USER_UUID" \
        "$REGISTRATION" \
        "$DISTANCE" \
        "$ADDITIONAL_TIRE_SIZES" \
        "$BODY_SHAPE" \
        "$CHASSIS_NUMBER" \
        "$CO2_EMISSIONS" \
        "$COLOUR" \
        "$DATE_FIRST_REGISTRATION" \
        "$EC_CATEGORY" \
        "$ENGINE_CAPACITY" \
        "$ENGINE_POWER" \
        "$NOISE_TEST_SPEED" \
        "$FIRST_REGISTRATION_IN_CROATIA" \
        "$FUEL_SOURCE" \
        "$HEIGHT" \
        "$HOMOLOGATION_TYPE" \
        "$LENGTH" \
        "$MARK" \
        "$MAX_NET_POWER" \
        "$MB_FIELD" \
        "$VEHICLE_MODEL" \
        "$NUM_AXLES" \
        "$NUM_DRIVEN_AXLES" \
        "$NUM_SEATS" \
        "$PERM_MAX_MASS" \
        "$PERM_PAYLOAD" \
        "$RATED_ENGINE_SPEED" \
        "$STATIONARY_NOISE" \
        "$TECH_MAX_MASS" \
        "$TIRE_SIZE" \
        "$TRADE_NAME" \
        "$TYPE_APPROVAL_NUMBER" \
        "$UNIQUE_MODEL_CODE" \
        "$UNLADEN_MASS" \
        "$VEHICLE_CATEGORY" \
        "$VEHICLE_TYPE" \
        "$VEHICLE_USE" \
        "$WIDTH"
) # End of printf arguments

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
