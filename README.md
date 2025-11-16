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

Configure OAuth2 clients and Redis storage in your configuration:

```yaml
redis:
  enabled: true
  host: "localhost"
  port: "6379"
  password: ""
  db: 0
```

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

### Obtaining a Token

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

### Using the Token

Include the token in API requests:

```bash
curl -X GET https://localhost:9443/api/users \
  -H "Authorization: Bearer eyJhbGc..."
```

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

- The self-signed certificates are for testing only
- Use proper CA-signed certificates in production
- Keep OAuth2 client secrets secure
- Use environment variables for sensitive configuration
- Enable Redis for production token storage
- Configure LDAP admin credentials securely

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

