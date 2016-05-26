package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	em "github.com/labstack/echo/middleware"

	// load app files
	"github.com/rightscale/azure_arm_proxy/config"
	eh "github.com/rightscale/azure_arm_proxy/error_handler"
	am "github.com/rightscale/azure_arm_proxy/middleware"
	"github.com/rightscale/azure_arm_proxy/resources"
)

func main() {
	// Serve
	s := httpServer()
	log.Printf("Azure plugin - listening on %s under %s environment\n", *config.ListenFlag, *config.Env)
	s.Run(*config.ListenFlag)
}

// Factory method for application
// Makes it possible to do integration testing.
func httpServer() *echo.Echo {

	// Setup middleware
	e := echo.New()
	e.Use(am.AzureClientInitializer())
	e.Use(em.Recover())

	if config.DebugMode {
		e.SetDebug(true)
	}

	e.SetHTTPErrorHandler(eh.AzureErrorHandler(e)) // override default error handler

	// Setup routes
	e.Get("/health-check", healthCheck)
	prefix := e.Group(*config.AppPrefix) // added prefix to use multiple nginx location on one SS box
	resources.SetupSubscriptionRoutes(prefix)
	resources.SetupInstanceRoutes(prefix)
	resources.SetupGroupsRoutes(prefix)
	resources.SetupStorageAccountsRoutes(prefix)
	resources.SetupProviderRoutes(prefix)
	resources.SetupNetworkRoutes(prefix)
	resources.SetupSubnetsRoutes(prefix)
	resources.SetupIPAddressesRoutes(prefix)
	resources.SetupAuthRoutes(prefix)
	resources.SetupNetworkInterfacesRoutes(prefix)
	resources.SetupImageRoutes(prefix)
	resources.SetupOperationRoutes(prefix)
	resources.SetupAvailabilitySetRoutes(prefix)
	resources.SetupNetworkSecurityGroupRoutes(prefix)
	resources.SetupNetworkSecurityGroupRuleRoutes(prefix)
	resources.SetupInstanceTypesRoutes(prefix)
	resources.SetupRouteTablesRoutes(prefix)
	resources.SetupRoutes(prefix)
	resources.SetupVirtualNetworkGatewayRoutes(prefix)
	resources.SetupEventsRoutes(prefix)

	return e
}

func healthCheck(c *echo.Context) error {
	return c.String(http.StatusOK, "Ok")
}
