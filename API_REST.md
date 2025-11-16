# OpenLDAP Node - REST API Documentation

## Overview

OpenLDAP Node provides a comprehensive REST API for managing users and groups in OpenLDAP with OAuth2 authentication. The API supports both user authentication (via LDAP credentials) and machine-to-machine authentication (client credentials flow).

## Base URL

```
http://localhost:9090
```

## Authentication

The API uses OAuth2 for authentication. All protected endpoints require a Bearer token in the Authorization header.

### OAuth2 Endpoints

#### POST /oauth/token

Obtain an access token using OAuth2.

**Grant Types Supported:**
- `password` - User authentication with LDAP credentials
- `client_credentials` - Machine-to-machine authentication

**Request (Password Grant):**
```bash
curl -X POST http://localhost:9090/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "username=john.doe" \
  -d "password=userpassword" \
  -d "client_id=openldap-client" \
  -d "client_secret=openldap-secret"
```

**Request (Client Credentials):**
```bash
curl -X POST http://localhost:9090/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=openldap-client" \
  -d "client_secret=openldap-secret"
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 7200,
  "refresh_token": "..."
}
```

#### POST /oauth/register

Register a new OAuth2 client (protected endpoint).

**Request:**
```bash
curl -X POST http://localhost:9090/oauth/register \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "my-app",
    "client_secret": "my-secret",
    "domain": "http://localhost:3000"
  }'
```

---

## User Management API

All user endpoints are protected and require OAuth2 authentication.

### GET /api/users

List all users.

**Request:**
```bash
curl -X GET http://localhost:9090/api/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "users": [
    {
      "dn": "uid=john.doe,ou=users,dc=example,dc=com",
      "username": "john.doe",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com",
      "display_name": "John Doe",
      "description": "Software Engineer"
    }
  ],
  "total": 1
}
```

### GET /api/users/search?q=query

Search users by username, name, or email.

**Request:**
```bash
curl -X GET "http://localhost:9090/api/users/search?q=john" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "users": [...],
  "total": 5,
  "query": "john"
}
```

### GET /api/users/:username

Get a specific user by username.

**Request:**
```bash
curl -X GET http://localhost:9090/api/users/john.doe \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "dn": "uid=john.doe,ou=users,dc=example,dc=com",
  "username": "john.doe",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "display_name": "John Doe",
  "description": "Software Engineer"
}
```

### POST /api/users

Create a new user.

**Request:**
```bash
curl -X POST http://localhost:9090/api/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "jane.smith",
    "password": "SecurePassword123!",
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane.smith@example.com",
    "display_name": "Jane Smith",
    "description": "DevOps Engineer"
  }'
```

**Response:**
```json
{
  "message": "User created successfully",
  "dn": "uid=jane.smith,ou=users,dc=example,dc=com"
}
```

### PUT /api/users/:username

Update an existing user.

**Request:**
```bash
curl -X PUT http://localhost:9090/api/users/jane.smith \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith-Johnson",
    "email": "jane.smith-johnson@example.com",
    "description": "Senior DevOps Engineer"
  }'
```

**Response:**
```json
{
  "message": "User updated successfully"
}
```

### DELETE /api/users/:username

Delete a user.

**Request:**
```bash
curl -X DELETE http://localhost:9090/api/users/jane.smith \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "message": "User deleted successfully"
}
```

---

## Group Management API

All group endpoints are protected and require OAuth2 authentication.

### GET /api/groups

List all groups.

**Request:**
```bash
curl -X GET http://localhost:9090/api/groups \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "groups": [
    {
      "dn": "cn=developers,ou=roles,dc=example,dc=com",
      "group_name": "developers",
      "description": "Development Team",
      "members": [
        "uid=john.doe,ou=users,dc=example,dc=com",
        "uid=jane.smith,ou=users,dc=example,dc=com"
      ],
      "gid_number": "50000"
    }
  ],
  "total": 1
}
```

### GET /api/groups/search?q=query

Search groups by name or description.

**Request:**
```bash
curl -X GET "http://localhost:9090/api/groups/search?q=dev" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### GET /api/groups/:groupname

Get a specific group.

**Request:**
```bash
curl -X GET http://localhost:9090/api/groups/developers \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "dn": "cn=developers,ou=roles,dc=example,dc=com",
  "group_name": "developers",
  "description": "Development Team",
  "members": [
    "uid=john.doe,ou=users,dc=example,dc=com"
  ],
  "gid_number": "50000"
}
```

### POST /api/groups

Create a new group.

**Request:**
```bash
curl -X POST http://localhost:9090/api/groups \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "devops",
    "description": "DevOps Team",
    "members": ["john.doe", "jane.smith"]
  }'
```

**Response:**
```json
{
  "message": "Group created successfully",
  "dn": "cn=devops,ou=roles,dc=example,dc=com"
}
```

### PUT /api/groups/:groupname

Update a group.

**Request:**
```bash
curl -X PUT http://localhost:9090/api/groups/devops \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "DevOps and SRE Team",
    "members": ["john.doe", "jane.smith", "bob.wilson"]
  }'
```

**Response:**
```json
{
  "message": "Group updated successfully"
}
```

### DELETE /api/groups/:groupname

Delete a group.

**Request:**
```bash
curl -X DELETE http://localhost:9090/api/groups/devops \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "message": "Group deleted successfully"
}
```

---

## Group Membership Management

### POST /api/groups/:groupname/members

Add a member to a group.

**Request:**
```bash
curl -X POST http://localhost:9090/api/groups/developers/members \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice.johnson"
  }'
```

**Response:**
```json
{
  "message": "Member added successfully"
}
```

### DELETE /api/groups/:groupname/members/:username

Remove a member from a group.

**Request:**
```bash
curl -X DELETE http://localhost:9090/api/groups/developers/members/alice.johnson \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "message": "Member removed successfully"
}
```

---

## Health & Utility Endpoints

### GET /api/health

Check API health status (public endpoint).

**Request:**
```bash
curl -X GET http://localhost:9090/api/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "0.1.3"
}
```

### GET /api/hello

Simple hello world endpoint (public endpoint).

**Request:**
```bash
curl -X GET http://localhost:9090/api/hello
```

**Response:**
```json
{
  "message": "world"
}
```

---

## Monitoring Endpoint

### GET /api/monitor/0

Get LDAP monitoring statistics (protected endpoint).

**Request:**
```bash
curl -X GET http://localhost:9090/api/monitor/0 \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "time": 1699876543,
  "legend": ["monitorOpInitiated", "monitorOpCompleted"],
  "value": [150, 148]
}
```

---

## CORS Configuration

The API supports CORS for the following origins:
- `http://localhost:3000`
- `http://localhost:9090`
- `http://127.0.0.1:3000`
- `http://127.0.0.1:9090`

Allowed methods: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`

---

## Error Handling

All endpoints return appropriate HTTP status codes:

- `200 OK` - Successful request
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

**Error Response Example:**
```json
{
  "error": "Invalid or expired token"
}
```

---

## Default OAuth2 Client

The server comes with a default OAuth2 client pre-configured:

- **Client ID:** `openldap-client`
- **Client Secret:** `openldap-secret`
- **Domain:** `http://localhost:9090`

---

## Security Notes

1. **Always use HTTPS in production** - The examples use HTTP for local development only
2. **Change default secrets** - Update the OAuth2 client secret and JWT signing key before deploying to production
3. **Token expiration** - Access tokens expire after 2 hours, refresh tokens after 7 days
4. **LDAP authentication** - User passwords are validated against LDAP using SSHA hashing
5. **Authorization header format** - `Authorization: Bearer <access_token>`

---

## Complete Workflow Example

### 1. Obtain Access Token
```bash
TOKEN=$(curl -s -X POST http://localhost:9090/oauth/token \
  -d "grant_type=password" \
  -d "username=admin" \
  -d "password=adminpass" \
  -d "client_id=openldap-client" \
  -d "client_secret=openldap-secret" \
  | jq -r '.access_token')
```

### 2. Create a User
```bash
curl -X POST http://localhost:9090/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john.doe",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com"
  }'
```

### 3. Create a Group
```bash
curl -X POST http://localhost:9090/api/groups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "developers",
    "description": "Development Team",
    "members": ["john.doe"]
  }'
```

### 4. List All Users
```bash
curl -X GET http://localhost:9090/api/users \
  -H "Authorization: Bearer $TOKEN"
```

---

## Support

For issues and questions, please refer to the project repository.
