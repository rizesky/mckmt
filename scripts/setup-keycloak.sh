#!/bin/bash

# Keycloak Setup Script for MCKMT
# This script sets up Keycloak with the necessary realm, client, and users for MCKMT

set -e

KEYCLOAK_URL="http://localhost:8082"
ADMIN_USER="admin"
ADMIN_PASSWORD="admin123"
REALM_NAME="mckmt"
CLIENT_ID="mckmt-hub"
CLIENT_SECRET="mckmt-client-secret-123"

echo "🔧 Setting up Keycloak for MCKMT..."

# Wait for Keycloak to be ready
echo "⏳ Waiting for Keycloak to be ready..."
until curl -s -f "${KEYCLOAK_URL}/realms/master" > /dev/null; do
    echo "   Keycloak is not ready yet, waiting..."
    sleep 5
done
echo "✅ Keycloak is ready!"

# Get admin access token
echo "🔑 Getting admin access token..."
ADMIN_TOKEN=$(curl -s -X POST "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=${ADMIN_USER}" \
    -d "password=${ADMIN_PASSWORD}" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" | jq -r '.access_token')

if [ "$ADMIN_TOKEN" = "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Failed to get admin token"
    exit 1
fi
echo "✅ Admin token obtained"

# Create realm
echo "🏰 Creating realm '${REALM_NAME}'..."
curl -s -X POST "${KEYCLOAK_URL}/admin/realms" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "realm": "'${REALM_NAME}'",
        "enabled": true,
        "displayName": "MCKMT Realm",
        "displayNameHtml": "<div class=\"kc-logo-text\"><span>MCKMT</span></div>",
        "loginWithEmailAllowed": true,
        "duplicateEmailsAllowed": false,
        "resetPasswordAllowed": true,
        "editUsernameAllowed": false,
        "bruteForceProtected": true,
        "permanentLockout": false,
        "maxFailureWaitSeconds": 900,
        "minimumQuickLoginWaitSeconds": 60,
        "waitIncrementSeconds": 60,
        "quickLoginCheckMilliSeconds": 1000,
        "maxDeltaTimeSeconds": 43200,
        "failureFactor": 30
    }' > /dev/null
echo "✅ Realm created"

# Create client
echo "🔐 Creating client '${CLIENT_ID}'..."
CLIENT_RESPONSE=$(curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/clients" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "clientId": "'${CLIENT_ID}'",
        "enabled": true,
        "clientAuthenticatorType": "client-secret",
        "secret": "'${CLIENT_SECRET}'",
        "redirectUris": ["http://localhost:8080/*", "http://localhost:3000/*"],
        "webOrigins": ["http://localhost:8080", "http://localhost:3000"],
        "protocol": "openid-connect",
        "publicClient": false,
        "serviceAccountsEnabled": true,
        "authorizationServicesEnabled": false,
        "standardFlowEnabled": true,
        "implicitFlowEnabled": false,
        "directAccessGrantsEnabled": true,
        "attributes": {
            "access.token.lifespan": "3600",
            "client.session.idle.timeout": "1800",
            "client.session.max.lifespan": "36000"
        }
    }')

if echo "$CLIENT_RESPONSE" | grep -q "error"; then
    echo "❌ Failed to create client: $CLIENT_RESPONSE"
    exit 1
fi
echo "✅ Client created"

# Get client UUID
CLIENT_UUID=$(curl -s -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/clients?clientId=${CLIENT_ID}" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq -r '.[0].id')

# Create roles
echo "👥 Creating roles..."
curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/roles" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"name": "admin", "description": "Administrator role with full access"}' > /dev/null

curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/roles" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"name": "operator", "description": "Operator role with cluster management access"}' > /dev/null

curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/roles" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"name": "viewer", "description": "Viewer role with read-only access"}' > /dev/null
echo "✅ Roles created"

# Create users
echo "👤 Creating users..."

# Admin user
curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "admin",
        "email": "admin@mckmt.local",
        "firstName": "Admin",
        "lastName": "User",
        "enabled": true,
        "emailVerified": true,
        "credentials": [{
            "type": "password",
            "value": "admin123",
            "temporary": false
        }]
    }' > /dev/null

# Get admin user ID
ADMIN_USER_ID=$(curl -s -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users?username=admin" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq -r '.[0].id')

# Assign admin role to admin user
curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users/${ADMIN_USER_ID}/role-mappings/realm" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '[{"id": "'$(curl -s -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/roles/admin" -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq -r '.id')'", "name": "admin"}]' > /dev/null

# Test user
curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "email": "test@mckmt.local",
        "firstName": "Test",
        "lastName": "User",
        "enabled": true,
        "emailVerified": true,
        "credentials": [{
            "type": "password",
            "value": "test123",
            "temporary": false
        }]
    }' > /dev/null

# Get test user ID
TEST_USER_ID=$(curl -s -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users?username=testuser" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq -r '.[0].id')

# Assign viewer role to test user
curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users/${TEST_USER_ID}/role-mappings/realm" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '[{"id": "'$(curl -s -X GET "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/roles/viewer" -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq -r '.id')'", "name": "viewer"}]' > /dev/null

echo "✅ Users created"

# Create OIDC configuration file
echo "📝 Creating OIDC configuration..."
mkdir -p configs/demo
cat > configs/demo/oidc-config.yaml << EOF
# OIDC Configuration for MCKMT Demo
# This configuration is specific to the demo environment with Keycloak

oidc:
  enabled: true
  issuer: "http://localhost:8082/realms/mckmt"
  client_id: "mckmt-hub"
  client_secret: "mckmt-client-secret-123"
  redirect_url: "http://localhost:8080/auth/oidc/callback"
  scopes: ["openid", "profile", "email"]
EOF

echo "✅ OIDC configuration created at configs/demo/oidc-config.yaml"

echo ""
echo "🎉 Keycloak setup completed!"
echo ""
echo "📋 Summary:"
echo "   • Keycloak URL: http://localhost:8082"
echo "   • Admin Console: http://localhost:8082/admin"
echo "   • Admin Username: admin"
echo "   • Admin Password: admin123"
echo "   • Realm: mckmt"
echo "   • Client ID: mckmt-hub"
echo "   • Client Secret: mckmt-client-secret-123"
echo ""
echo "👥 Test Users:"
echo "   • admin@mckmt.local / admin123 (admin role)"
echo "   • test@mckmt.local / test123 (viewer role)"
echo ""
echo "🔧 Next steps:"
echo "   1. Update your hub.yaml to enable OIDC"
echo "   2. Set the OIDC configuration values"
echo "   3. Restart the hub service"
echo ""
