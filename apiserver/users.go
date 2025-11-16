package apiserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aescanero/openldap-node/ldaputils"
	"github.com/aescanero/openldap-node/utils"
	"github.com/gin-gonic/gin"
	ldap "github.com/go-ldap/ldap/v3"
)

// UserRequest represents a user creation/update request
type UserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// UserResponse represents a user in responses
type UserResponse struct {
	DN          string `json:"dn"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// getAdminConnection gets an LDAP admin connection
func getAdminConnection() (*ldap.Conn, error) {
	adminPassword, err := apiConfig.SrvConfig.GetAdminPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to get admin password: %w", err)
	}

	adminDN := fmt.Sprintf("cn=admin,%s", apiConfig.Database[0].Base)
	conn, err := ldaputils.Connect(apiConfig, adminDN, adminPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to connect as admin: %w", err)
	}

	return conn, nil
}

// ListUsers returns users from LDAP with pagination support
func ListUsers(c *gin.Context) {
	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	// Parse pagination parameters
	page, pageSize := getPaginationParams(c)

	baseDN := fmt.Sprintf("ou=users,%s", apiConfig.Database[0].Base)
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=inetOrgPerson)",
		[]string{"uid", "cn", "sn", "givenName", "mail", "displayName", "description"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search failed: %v", err)})
		return
	}

	total := len(result.Entries)

	// Apply pagination
	start, end := calculatePagination(page, pageSize, total)
	paginatedEntries := result.Entries[start:end]

	users := make([]UserResponse, 0, len(paginatedEntries))
	for _, entry := range paginatedEntries {
		users = append(users, UserResponse{
			DN:          entry.DN,
			Username:    entry.GetAttributeValue("uid"),
			FirstName:   entry.GetAttributeValue("givenName"),
			LastName:    entry.GetAttributeValue("sn"),
			Email:       entry.GetAttributeValue("mail"),
			DisplayName: entry.GetAttributeValue("displayName"),
			Description: entry.GetAttributeValue("description"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// GetUser returns a specific user by username
func GetUser(c *gin.Context) {
	username := c.Param("username")

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	userDN := fmt.Sprintf("uid=%s,ou=users,%s", username, apiConfig.Database[0].Base)
	searchRequest := ldap.NewSearchRequest(
		userDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=inetOrgPerson)",
		[]string{"uid", "cn", "sn", "givenName", "mail", "displayName", "description"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if len(result.Entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	entry := result.Entries[0]
	user := UserResponse{
		DN:          entry.DN,
		Username:    entry.GetAttributeValue("uid"),
		FirstName:   entry.GetAttributeValue("givenName"),
		LastName:    entry.GetAttributeValue("sn"),
		Email:       entry.GetAttributeValue("mail"),
		DisplayName: entry.GetAttributeValue("displayName"),
		Description: entry.GetAttributeValue("description"),
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user in LDAP
func CreateUser(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	// Generate password hash
	encode := utils.Encode{}
	passwordHash := encode.MakeSSHAEncode([]byte(req.Password))
	passwordSSHA := fmt.Sprintf("{SSHA}%s", utils.EncodeBase64(passwordHash))

	// Build CN from first and last name, or use username
	cn := req.Username
	if req.FirstName != "" && req.LastName != "" {
		cn = fmt.Sprintf("%s %s", req.FirstName, req.LastName)
	} else if req.DisplayName != "" {
		cn = req.DisplayName
	}

	// Build user DN
	userDN := fmt.Sprintf("uid=%s,ou=users,%s", req.Username, apiConfig.Database[0].Base)

	// Create add request
	addRequest := ldap.NewAddRequest(userDN, nil)
	addRequest.Attribute("objectClass", []string{"inetOrgPerson", "posixAccount", "top"})
	addRequest.Attribute("uid", []string{req.Username})
	addRequest.Attribute("cn", []string{cn})
	addRequest.Attribute("sn", []string{req.LastName})
	addRequest.Attribute("userPassword", []string{passwordSSHA})
	addRequest.Attribute("uidNumber", []string{fmt.Sprintf("%d", 10000)}) // TODO: Auto-increment
	addRequest.Attribute("gidNumber", []string{"10000"})
	addRequest.Attribute("homeDirectory", []string{fmt.Sprintf("/home/%s", req.Username)})

	if req.FirstName != "" {
		addRequest.Attribute("givenName", []string{req.FirstName})
	}
	if req.Email != "" {
		addRequest.Attribute("mail", []string{req.Email})
	}
	if req.DisplayName != "" {
		addRequest.Attribute("displayName", []string{req.DisplayName})
	}
	if req.Description != "" {
		addRequest.Attribute("description", []string{req.Description})
	}

	// Execute add
	err = conn.Add(addRequest)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create user: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"dn":      userDN,
	})
}

// UpdateUser updates an existing user in LDAP
func UpdateUser(c *gin.Context) {
	username := c.Param("username")

	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	userDN := fmt.Sprintf("uid=%s,ou=users,%s", username, apiConfig.Database[0].Base)

	// Build modify request
	modifyRequest := ldap.NewModifyRequest(userDN, nil)

	if req.FirstName != "" {
		modifyRequest.Replace("givenName", []string{req.FirstName})
	}
	if req.LastName != "" {
		modifyRequest.Replace("sn", []string{req.LastName})
	}
	if req.Email != "" {
		modifyRequest.Replace("mail", []string{req.Email})
	}
	if req.DisplayName != "" {
		modifyRequest.Replace("displayName", []string{req.DisplayName})
	}
	if req.Description != "" {
		modifyRequest.Replace("description", []string{req.Description})
	}

	// Update password if provided
	if req.Password != "" {
		encode := utils.Encode{}
		passwordHash := encode.MakeSSHAEncode([]byte(req.Password))
		passwordSSHA := fmt.Sprintf("{SSHA}%s", utils.EncodeBase64(passwordHash))
		modifyRequest.Replace("userPassword", []string{passwordSSHA})
	}

	// Update CN if first or last name changed
	if req.FirstName != "" && req.LastName != "" {
		cn := fmt.Sprintf("%s %s", req.FirstName, req.LastName)
		modifyRequest.Replace("cn", []string{cn})
	}

	err = conn.Modify(modifyRequest)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser deletes a user from LDAP
func DeleteUser(c *gin.Context) {
	username := c.Param("username")

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	userDN := fmt.Sprintf("uid=%s,ou=users,%s", username, apiConfig.Database[0].Base)

	// Create delete request
	delRequest := ldap.NewDelRequest(userDN, nil)

	err = conn.Del(delRequest)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// SearchUsers searches for users by filter
func SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		ListUsers(c)
		return
	}

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	baseDN := fmt.Sprintf("ou=users,%s", apiConfig.Database[0].Base)

	// Build search filter for multiple attributes
	filter := fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=*%s*)(cn=*%s*)(mail=*%s*)))", query, query, query)

	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{"uid", "cn", "sn", "givenName", "mail", "displayName", "description"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search failed: %v", err)})
		return
	}

	users := make([]UserResponse, 0, len(result.Entries))
	for _, entry := range result.Entries {
		users = append(users, UserResponse{
			DN:          entry.DN,
			Username:    entry.GetAttributeValue("uid"),
			FirstName:   entry.GetAttributeValue("givenName"),
			LastName:    entry.GetAttributeValue("sn"),
			Email:       entry.GetAttributeValue("mail"),
			DisplayName: entry.GetAttributeValue("displayName"),
			Description: entry.GetAttributeValue("description"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "total": len(users), "query": query})
}

// getPaginationParams extracts and validates pagination parameters from request
func getPaginationParams(c *gin.Context) (page int, pageSize int) {
	page = 1
	pageSize = 20 // Default page size

	if p := c.Query("page"); p != "" {
		if parsed, err := fmt.Sscanf(p, "%d", &page); err == nil && parsed == 1 && page > 0 {
			// page is valid
		} else {
			page = 1
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := fmt.Sscanf(ps, "%d", &pageSize); err == nil && parsed == 1 && pageSize > 0 {
			// pageSize is valid
			if pageSize > 100 {
				pageSize = 100 // Max page size limit
			}
		} else {
			pageSize = 20
		}
	}

	return page, pageSize
}

// calculatePagination calculates start and end indices for pagination
func calculatePagination(page, pageSize, total int) (start int, end int) {
	start = (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= total {
		start = 0
		if total > 0 {
			start = total
		}
	}

	end = start + pageSize
	if end > total {
		end = total
	}

	return start, end
}
