package main

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	docs "github.com/habibiefaried/whatsapp-sender/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// MessageRequest represents the structure of the request body
//
//	@Description The message request body
//	@Param message body MessageRequest true "Input message"
//	@Example {"message": "Hello, World!"}
type MessageRequest struct {
	Message string `json:"message"`
}

// Credentials structure to hold username and password
type Credentials struct {
	Username string
	Password string
}

// Global variable to hold credentials
var validCredentials Credentials

// LoadCredentials reads the username and password from a file
func LoadCredentials(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Assuming credentials are in "username:password" format
	parts := strings.TrimSpace(string(data))
	credentialParts := strings.Split(parts, ":")
	if len(credentialParts) != 2 {
		return err
	}

	validCredentials = Credentials{
		Username: credentialParts[0],
		Password: credentialParts[1],
	}
	return nil
}

// @BasePath /api/v1

// @securityDefinitions.basic BasicAuth
// @Description Basic Auth
// @Name Authorization
// @In header
// @Type basic
// @Title Basic Auth

// @Summary send message
// @Schemes
// @Description send message
// @Accept json
// @Produce json
// @Param request body MessageRequest true "Input message"
// @Success 200 {object} map[string]string
// @Router /sendMessage [post]
// @Security BasicAuth
func SendMessage(g *gin.Context) {
	// Check Authorization header
	if !validateBasicAuth(g) {
		return
	}

	var jsonBody MessageRequest

	if err := g.ShouldBindJSON(&jsonBody); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, gin.H{"response": "Received: " + jsonBody.Message})
}

// @Summary Receive message
// @Schemes
// @Description Receive a message with the specified number
// @Accept json
// @Produce json
// @Param number query string true "The number parameter"
// @Success 200 {object} map[string]string
// @Router /recvMessage [get]
// @Security BasicAuth
func RecvMessage(g *gin.Context) {
	// Check Authorization header
	if !validateBasicAuth(g) {
		return
	}

	number := g.Query("number") // Retrieve the "number" query parameter

	if number == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "number parameter is required"})
		return
	}

	g.JSON(http.StatusOK, gin.H{"message": "Received number: " + number})
}

// validateBasicAuth checks the Authorization header for Basic Auth
func validateBasicAuth(g *gin.Context) bool {
	authHeader := g.GetHeader("Authorization")
	if authHeader == "" {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return false
	}

	// Extract the token from the header
	token := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
		return false
	}

	// Split the decoded string into username and password
	credentials := strings.Split(string(decoded), ":")
	if len(credentials) != 2 || !validateCredentials(credentials[0], credentials[1]) {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return false
	}

	return true
}

// validateCredentials checks the provided username and password against the loaded credentials
func validateCredentials(username, password string) bool {
	return username == validCredentials.Username && password == validCredentials.Password
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	// Load credentials from file
	err := LoadCredentials("credentials.txt")
	if err != nil {
		panic("Failed to load credentials: " + err.Error())
	}

	r := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	v1.POST("/sendMessage", SendMessage)
	v1.GET("/recvMessage", RecvMessage)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Run(":45981")
}
