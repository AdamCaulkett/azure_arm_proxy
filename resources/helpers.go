package resources

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	eh "github.com/rightscale/azure_arm_proxy/error_handler"
	"github.com/rightscale/azure_arm_proxy/middleware"
)

// GetAzureClient retrieves client initialized by middleware, send error response if not found
// This function should be used by controller actions that need to use the client
func GetAzureClient(c *echo.Context) (*http.Client, error) {
	client, _ := c.Get("azure").(*http.Client)
	if client == nil {
		return nil, eh.GenericException(fmt.Sprintf("failed to retrieve Azure client, check middleware"))
	}
	return client, nil
}

// GetClientCredentials retrieves client credentials initialized by middleware, send error response if not found
// This function should be used by controller actions that need to operate with client credentials
func GetClientCredentials(c *echo.Context) (*middleware.Credentials, error) {
	creds, _ := c.Get("clientCreds").(*middleware.Credentials)
	if creds == nil {
		return nil, eh.GenericException(fmt.Sprintf("failed to retrieve client credentials, check middleware"))
	}
	return creds, nil
}
