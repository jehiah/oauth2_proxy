# Bearer Token Authentication

## Overview

Bearer token authentication allows API/CLI access to oauth2_proxy using Google ID tokens. This enables programmatic access while maintaining the same email and group restrictions as cookie-based authentication.

## Configuration

Enable bearer token authentication with these command-line flags:

```bash
oauth2_proxy \
  --provider=google \
  --client-id=YOUR_CLIENT_ID.apps.googleusercontent.com \
  --client-secret=YOUR_SECRET \
  --enable-bearer-token-auth \
  --cookie-secret=RANDOM_SECRET \
  --email-domain=example.com
```

### Configuration Flags

- `--enable-bearer-token-auth` (boolean, default: false) - Enables bearer token authentication
- `--google-id-token-audience` (repeatable string) - OAuth client IDs to accept in the token's `aud` claim. Defaults to the value of `--client-id` if not specified.

## Usage

### Generate a Google ID Token

**Important**: You must specify your OAuth client ID as the audience:

```bash
# Generate a token by impersonating a service account
TOKEN=$(gcloud auth print-identity-token \
  --audiences=YOUR_CLIENT_ID.apps.googleusercontent.com \
  --include-email \
  --impersonate-service-account=your-service-account@PROJECT_ID.iam.gserviceaccount.com)

# Or generate a token for your own credentials
TOKEN=$(gcloud auth print-identity-token)
```

**Note**: Without `--audiences`, gcloud generates a token for its own client ID, which will fail audience validation in oauth2_proxy. When impersonating a service account, `--include-email` is required so the token includes the `email` claim.

### Make API Requests

```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://oauth2-proxy.example.com/api/endpoint
```

### With Google Workspace Groups

Bearer token authentication respects Google Workspace group restrictions:

```bash
oauth2_proxy \
  --provider=google \
  --client-id=YOUR_CLIENT_ID \
  --enable-bearer-token-auth \
  --google-group=engineering@example.com \
  --google-admin-email=admin@example.com \
  --google-service-account-json=/path/to/credentials.json
```

## Authentication Flow

oauth2_proxy tries authentication methods in this order:

1. **Cookie** - Web-based session cookie
2. **Basic Auth** - Username/password from htpasswd file (if configured)
3. **Bearer Token** - Google ID token (if `--enable-bearer-token-auth` is enabled)

This cascade allows both web users and API clients to access the same protected resources.

## Security

- **Cryptographic verification**: ID tokens are verified using Google's public keys (automatically cached)
- **Audience validation**: The token's `aud` claim must match a configured client ID
- **Email verification**: Only tokens with `email_verified: true` are accepted
- **Expiration enforcement**: Tokens are typically valid for 1 hour
- **Group validation**: Google Workspace group membership is checked on every request
- **No token caching**: Per-request validation ensures immediate enforcement of access changes

## Headers Forwarded to Upstream

When authentication succeeds, oauth2_proxy forwards these headers to the upstream application:

- `X-Forwarded-User` - Username (email prefix)
- `X-Forwarded-Email` - Full email address

## Troubleshooting

### "bearer token auth only supported with Google provider"

Bearer token authentication currently only works with `--provider=google`.

### "invalid ID token: oidc: expected audience"

The token's `aud` claim doesn't match the configured client ID. Make sure you're generating the token with the correct audience:

```bash
# Specify your client ID as the audience
gcloud auth print-identity-token \
  --audiences=YOUR_CLIENT_ID.apps.googleusercontent.com
```

### "email not verified"

The Google account's email address must be verified. Check the token claims:

```bash
TOKEN=$(gcloud auth print-identity-token --audiences=YOUR_CLIENT_ID.apps.googleusercontent.com)
echo $TOKEN | cut -d. -f2 | base64 -d | jq .
```

Look for `"email_verified": true`.

### "email not authorized" or "not in allowed Google group"

The email must pass the same validation as cookie-based authentication:
- Match `--email-domain` or be in `--authenticated-emails-file`
- Be a member of `--google-group` (if configured)

### Debugging Tokens

Decode a Google ID token to inspect its claims:

```bash
# View header
echo $TOKEN | cut -d. -f1 | base64 -d | jq .

# View payload
echo $TOKEN | cut -d. -f2 | base64 -d | jq .
```

Expected payload structure:
```json
{
  "iss": "https://accounts.google.com",
  "aud": "YOUR_CLIENT_ID.apps.googleusercontent.com",
  "sub": "...",
  "email": "user@example.com",
  "email_verified": true,
  "exp": 1234567890,
  "iat": 1234567890
}
```

## Example: Python API Client

```python
import subprocess
import requests

CLIENT_ID = "YOUR_CLIENT_ID.apps.googleusercontent.com"

def get_id_token(audience):
    """Get Google ID token using gcloud."""
    result = subprocess.run(
        ['gcloud', 'auth', 'print-identity-token', f'--audiences={audience}'],
        capture_output=True,
        text=True,
        check=True
    )
    return result.stdout.strip()

# Make authenticated request
token = get_id_token(CLIENT_ID)
headers = {'Authorization': f'Bearer {token}'}
response = requests.get('https://oauth2-proxy.example.com/api/data', headers=headers)
print(response.json())
```

## Performance

- **ID token verification**: ~1-2ms with cached public keys
- **Group validation**: ~100-200ms per Admin SDK call (recommend caching in future)
- **No additional HTTP calls**: ID tokens are verified cryptographically (unlike access tokens which require HTTP validation)
