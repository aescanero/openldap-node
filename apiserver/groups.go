package apiserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ldap "github.com/go-ldap/ldap/v3"
)

// GroupRequest represents a group creation/update request
type GroupRequest struct {
	GroupName   string   `json:"group_name" binding:"required"`
	Description string   `json:"description"`
	Members     []string `json:"members"` // Array of usernames
}

// GroupResponse represents a group in responses
type GroupResponse struct {
	DN          string   `json:"dn"`
	GroupName   string   `json:"group_name"`
	Description string   `json:"description,omitempty"`
	Members     []string `json:"members,omitempty"`
	GIDNumber   string   `json:"gid_number,omitempty"`
}

// ListGroups returns all groups from LDAP
func ListGroups(c *gin.Context) {
	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	baseDN := fmt.Sprintf("ou=roles,%s", apiConfig.Database[0].Base)
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=groupOfNames)",
		[]string{"cn", "description", "member", "gidNumber"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search failed: %v", err)})
		return
	}

	groups := make([]GroupResponse, 0, len(result.Entries))
	for _, entry := range result.Entries {
		groups = append(groups, GroupResponse{
			DN:          entry.DN,
			GroupName:   entry.GetAttributeValue("cn"),
			Description: entry.GetAttributeValue("description"),
			Members:     entry.GetAttributeValues("member"),
			GIDNumber:   entry.GetAttributeValue("gidNumber"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups, "total": len(groups)})
}

// GetGroup returns a specific group by name
func GetGroup(c *gin.Context) {
	groupName := c.Param("groupname")

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	groupDN := fmt.Sprintf("cn=%s,ou=roles,%s", groupName, apiConfig.Database[0].Base)
	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=groupOfNames)",
		[]string{"cn", "description", "member", "gidNumber"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if len(result.Entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	entry := result.Entries[0]
	group := GroupResponse{
		DN:          entry.DN,
		GroupName:   entry.GetAttributeValue("cn"),
		Description: entry.GetAttributeValue("description"),
		Members:     entry.GetAttributeValues("member"),
		GIDNumber:   entry.GetAttributeValue("gidNumber"),
	}

	c.JSON(http.StatusOK, group)
}

// CreateGroup creates a new group in LDAP
func CreateGroup(c *gin.Context) {
	var req GroupRequest
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

	// Build group DN
	groupDN := fmt.Sprintf("cn=%s,ou=roles,%s", req.GroupName, apiConfig.Database[0].Base)

	// Build member DNs
	memberDNs := make([]string, len(req.Members))
	for i, username := range req.Members {
		memberDNs[i] = fmt.Sprintf("uid=%s,ou=users,%s", username, apiConfig.Database[0].Base)
	}

	// If no members, add a placeholder (required by groupOfNames)
	if len(memberDNs) == 0 {
		memberDNs = []string{fmt.Sprintf("cn=admin,%s", apiConfig.Database[0].Base)}
	}

	// Create add request
	addRequest := ldap.NewAddRequest(groupDN, nil)
	addRequest.Attribute("objectClass", []string{"groupOfNames", "posixGroup", "top"})
	addRequest.Attribute("cn", []string{req.GroupName})
	addRequest.Attribute("member", memberDNs)
	addRequest.Attribute("gidNumber", []string{fmt.Sprintf("%d", 50000)}) // TODO: Auto-increment

	if req.Description != "" {
		addRequest.Attribute("description", []string{req.Description})
	}

	// Execute add
	err = conn.Add(addRequest)
	if err != nil {
		log.Printf("Failed to create group: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create group: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Group created successfully",
		"dn":      groupDN,
	})
}

// UpdateGroup updates an existing group in LDAP
func UpdateGroup(c *gin.Context) {
	groupName := c.Param("groupname")

	var req GroupRequest
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

	groupDN := fmt.Sprintf("cn=%s,ou=roles,%s", groupName, apiConfig.Database[0].Base)

	// Build modify request
	modifyRequest := ldap.NewModifyRequest(groupDN, nil)

	if req.Description != "" {
		modifyRequest.Replace("description", []string{req.Description})
	}

	// Update members if provided
	if len(req.Members) > 0 {
		memberDNs := make([]string, len(req.Members))
		for i, username := range req.Members {
			memberDNs[i] = fmt.Sprintf("uid=%s,ou=users,%s", username, apiConfig.Database[0].Base)
		}
		modifyRequest.Replace("member", memberDNs)
	}

	err = conn.Modify(modifyRequest)
	if err != nil {
		log.Printf("Failed to update group: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update group: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group updated successfully"})
}

// DeleteGroup deletes a group from LDAP
func DeleteGroup(c *gin.Context) {
	groupName := c.Param("groupname")

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	groupDN := fmt.Sprintf("cn=%s,ou=roles,%s", groupName, apiConfig.Database[0].Base)

	// Create delete request
	delRequest := ldap.NewDelRequest(groupDN, nil)

	err = conn.Del(delRequest)
	if err != nil {
		log.Printf("Failed to delete group: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete group: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// AddMemberToGroup adds a user to a group
func AddMemberToGroup(c *gin.Context) {
	groupName := c.Param("groupname")

	var req struct {
		Username string `json:"username" binding:"required"`
	}
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

	groupDN := fmt.Sprintf("cn=%s,ou=roles,%s", groupName, apiConfig.Database[0].Base)
	userDN := fmt.Sprintf("uid=%s,ou=users,%s", req.Username, apiConfig.Database[0].Base)

	// Add member
	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Add("member", []string{userDN})

	err = conn.Modify(modifyRequest)
	if err != nil {
		log.Printf("Failed to add member to group: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add member: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member added successfully"})
}

// RemoveMemberFromGroup removes a user from a group
func RemoveMemberFromGroup(c *gin.Context) {
	groupName := c.Param("groupname")
	username := c.Param("username")

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	groupDN := fmt.Sprintf("cn=%s,ou=roles,%s", groupName, apiConfig.Database[0].Base)
	userDN := fmt.Sprintf("uid=%s,ou=users,%s", username, apiConfig.Database[0].Base)

	// Remove member
	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Delete("member", []string{userDN})

	err = conn.Modify(modifyRequest)
	if err != nil {
		log.Printf("Failed to remove member from group: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to remove member: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// SearchGroups searches for groups by filter
func SearchGroups(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		ListGroups(c)
		return
	}

	conn, err := getAdminConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	baseDN := fmt.Sprintf("ou=roles,%s", apiConfig.Database[0].Base)

	// Build search filter
	filter := fmt.Sprintf("(&(objectClass=groupOfNames)(|(cn=*%s*)(description=*%s*)))", query, query)

	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{"cn", "description", "member", "gidNumber"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search failed: %v", err)})
		return
	}

	groups := make([]GroupResponse, 0, len(result.Entries))
	for _, entry := range result.Entries {
		groups = append(groups, GroupResponse{
			DN:          entry.DN,
			GroupName:   entry.GetAttributeValue("cn"),
			Description: entry.GetAttributeValue("description"),
			Members:     entry.GetAttributeValues("member"),
			GIDNumber:   entry.GetAttributeValue("gidNumber"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups, "total": len(groups), "query": query})
}
