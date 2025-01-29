#!/bin/bash

set -exo pipefail

 # CI test for use with CI AWS account
if [ -z "$BATON_CONNECTOR" ]; then
  echo "BATON_CONNECTOR not set."
  exit 1
fi
if [ -z "$BATON" ]; then
  echo "BATON not set. using baton"
  BATON=baton
fi

# Error on unbound variables now that we've set BATON & BATON_CONNECTOR
set -u

# Sync
$BATON_CONNECTOR

# TODO: Uncomment these lines once provisioning is supported

# Grant entitlement
#$BATON_CONNECTOR --grant-entitlement="$CONNECTOR_ENTITLEMENT" --grant-principal="$CONNECTOR_PRINCIPAL" --grant-principal-type="$CONNECTOR_PRINCIPAL_TYPE"

# Check for grant before revoking
$BATON_CONNECTOR
$BATON grants --entitlement="$CONNECTOR_ENTITLEMENT" --output-format=json | jq --exit-status ".grants[] | select( .principal.id.resource == \"$CONNECTOR_PRINCIPAL\" )"

# Grant already-granted entitlement
#$BATON_CONNECTOR --grant-entitlement="$CONNECTOR_ENTITLEMENT" --grant-principal="$CONNECTOR_PRINCIPAL" --grant-principal-type="$CONNECTOR_PRINCIPAL_TYPE"

# Get grant ID
# CONNECTOR_GRANT=$($BATON grants --entitlement="$CONNECTOR_ENTITLEMENT" --output-format=json | jq --raw-output --exit-status ".grants[] | select( .principal.id.resource == \"$CONNECTOR_PRINCIPAL\" ).grant.id")

# Revoke grant
# $BATON_CONNECTOR --revoke-grant="$CONNECTOR_GRANT"

# Revoke already-revoked grant
# $BATON_CONNECTOR --revoke-grant="$CONNECTOR_GRANT"

# Check grant was revoked
# $BATON_CONNECTOR
# $BATON grants --entitlement="$CONNECTOR_ENTITLEMENT" --output-format=json | jq --exit-status "if .grants then [ .grants[] | select( .principal.id.resource == \"$CONNECTOR_PRINCIPAL\" ) ] | length == 0 else . end"

# Re-grant entitlement
# $BATON_CONNECTOR --grant-entitlement="$CONNECTOR_ENTITLEMENT" --grant-principal="$CONNECTOR_PRINCIPAL" --grant-principal-type="$CONNECTOR_PRINCIPAL_TYPE"

# Check grant was re-granted
# $BATON_CONNECTOR
# $BATON grants --entitlement="$CONNECTOR_ENTITLEMENT" --output-format=json | jq --exit-status ".grants[] | select( .principal.id.resource == \"$CONNECTOR_PRINCIPAL\" )"
