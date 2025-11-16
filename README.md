# openldap-node
A OpenLDAP Server managed by a Go controller with REST API and OAuth2 authentication

## Features

- **OpenLDAP Integration**: Full LDAP server management
- **REST API**: Complete CRUD operations for users and groups
- **OAuth2 Server**: Built-in OAuth2 server with LDAP authentication backend
  - Password grant flow (user authentication)
  - Client credentials grant (machine-to-machine)
  - JWT token-based authentication
  - Redis token storage (with memory fallback)
- **HTTPS/TLS Support**: Secure communication with certificate-based encryption
- **React Admin Dashboard**: Modern web UI for user and group management
- **CORS Support**: Cross-origin resource sharing enabled
- **Pagination**: Efficient data retrieval with pagination support

## Architecture

- **Backend**: Go 1.23 with Gin web framework
- **Authentication**: OAuth2 server with LDAP backend
- **Frontend**: React with Material-UI and React Admin
- **LDAP**: OpenLDAP server integration
- **Token Storage**: Redis (optional, with memory fallback)

## Installation

### Prerequisites

- Go 1.23 or later
- Node.js and npm (for frontend development)
- OpenLDAP server
- Redis server (optional, for token persistence)

### Build

```bash
# Build the Go server
go build -o openldap-node

# Build the frontend (optional)
cd apiserver/dashboard
npm install
npm run build
cd ../..
```

## Configuration

All configuration can be provided via YAML configuration file or environment variables. Environment variables take precedence over configuration file values.

### Environment Variables

- `OAUTH2_ENABLED`: Enable/disable OAuth2 authentication (`true` or `false`)
- `REDIS_ENABLED`: Enable/disable Redis token storage (`true` or `false`)
- `REDIS_HOST`: Redis server host
- `REDIS_PORT`: Redis server port
- `REDIS_PASSWORD`: Redis password
- `REDIS_DB`: Redis database number
- `API_TLS_ENABLED`: Enable/disable HTTPS (`true` or `false`)
- `API_TLS_CERT_FILE`: Path to TLS certificate file
- `API_TLS_KEY_FILE`: Path to TLS key file
- `API_TLS_PORT`: HTTPS port (default: 9443)
- `API_TLS_AUTO_REDIRECT`: Auto-redirect HTTP to HTTPS (`true` or `false`)

Example:
```bash
export OAUTH2_ENABLED=false
export API_TLS_ENABLED=true
./openldap-node server
```

### HTTPS/TLS

Generate self-signed certificates for testing:

```bash
bash scripts/generate-certs.sh
```

Configure TLS in your configuration file:

```yaml
api_tls:
  enabled: true
  cert_file: "certs/server.crt"
  key_file: "certs/server.key"
  port: ":9443"
  auto_redirect: true  # Redirect HTTP to HTTPS
```

### OAuth2

The API supports two authentication modes:

1. **OAuth2 Enabled** (default): API endpoints require OAuth2 Bearer token authentication
2. **OAuth2 Disabled**: API endpoints are publicly accessible without authentication

Configure OAuth2 mode and Redis storage in your configuration:

```yaml
oauth2:
  enabled: true  # Set to false to disable authentication

redis:
  enabled: true
  host: "localhost"
  port: "6379"
  password: ""
  db: 0
```

**Important**: When OAuth2 is disabled, all API endpoints become publicly accessible. Use this mode only in trusted networks or for development purposes.

### LDAP

Configure LDAP connection settings:

```yaml
database:
  - base: "dc=example,dc=com"
    suffix: "dc=example,dc=com"
    rootdn: "cn=admin,dc=example,dc=com"
    rootpw: "secret"
```

## API Endpoints

### OAuth2

- `POST /oauth/token` - Obtain access token
  - Password grant: username/password authentication
  - Client credentials: machine-to-machine authentication
- `POST /oauth/authorize` - Authorization endpoint

### Users API

- `GET /api/users` - List users (with pagination)
- `GET /api/users/:username` - Get user details
- `POST /api/users` - Create new user
- `PUT /api/users/:username` - Update user
- `DELETE /api/users/:username` - Delete user
- `GET /api/users/search?q=query` - Search users

### Groups API

- `GET /api/groups` - List groups (with pagination)
- `GET /api/groups/:groupname` - Get group details
- `POST /api/groups` - Create new group
- `PUT /api/groups/:groupname` - Update group
- `DELETE /api/groups/:groupname` - Delete group
- `POST /api/groups/:groupname/members` - Add member to group
- `DELETE /api/groups/:groupname/members/:username` - Remove member from group
- `GET /api/groups/search?q=query` - Search groups

### Pagination

All list endpoints support pagination:
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)

Example:
```bash
curl "https://localhost:9443/api/users?page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

## Authentication

The server supports two authentication modes:

### Mode 1: OAuth2 Enabled (Default)

When OAuth2 is enabled (`oauth2.enabled: true`), all API endpoints require authentication.

#### Obtaining a Token

```bash
curl -X POST https://localhost:9443/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "username=admin" \
  -d "password=secret" \
  -d "client_id=openldap-client" \
  -d "client_secret=openldap-secret"
```

Response:
```json
{
  "access_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 7200
}
```

#### Using the Token

Include the token in API requests:

```bash
curl -X GET https://localhost:9443/api/users \
  -H "Authorization: Bearer eyJhbGc..."
```

### Mode 2: OAuth2 Disabled

When OAuth2 is disabled (`oauth2.enabled: false`), all API endpoints are publicly accessible without authentication.

#### Direct API Access

```bash
# No authentication required
curl -X GET https://localhost:9443/api/users

# Create a user without token
curl -X POST https://localhost:9443/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "givenName": "John",
    "sn": "Doe",
    "password": "userpass123"
  }'
```

**Security Warning**: Only use disabled mode in trusted networks or for development. In production, always enable OAuth2 authentication.

## React Admin Dashboard

The project includes a React-based admin dashboard for managing users and groups.

### Development

```bash
cd apiserver/dashboard
npm start
```

Access the dashboard at http://localhost:3000

### Production

```bash
cd apiserver/dashboard
npm run build
```

The built files will be in the `build/` directory and can be served by the Go server or any static file server.

### Features

- User management (list, create, edit, delete)
- Group management (list, create, edit, delete, manage members)
- OAuth2 authentication
- Search and filtering
- Pagination
- Export functionality

## Running the Server

### HTTP Mode

```bash
./openldap-node server
```

### HTTPS Mode

```bash
# Generate certificates first
bash scripts/generate-certs.sh

# Run with TLS enabled
./openldap-node server
```

The server will:
- Listen on HTTP port 9090 (if configured)
- Listen on HTTPS port 9443 (if TLS enabled)
- Auto-redirect HTTP to HTTPS (if auto_redirect is enabled)

## Security Notes

- **OAuth2 Authentication**: Always enable OAuth2 (`oauth2.enabled: true`) in production environments
- **Disabled Mode**: Only use OAuth2 disabled mode in trusted networks or development environments
- **TLS Certificates**: The self-signed certificates are for testing only; use proper CA-signed certificates in production
- **Client Secrets**: Keep OAuth2 client secrets secure and never commit them to version control
- **Environment Variables**: Use environment variables for sensitive configuration
- **Redis**: Enable Redis for production token storage to support distributed deployments
- **LDAP Credentials**: Configure LDAP admin credentials securely using password files or environment variables
- **Network Security**: When OAuth2 is disabled, ensure the API is not exposed to untrusted networks

## License

Copyright [2023] [Alejandro Escanero Blanco <aescanero@disasterproject.com>]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

