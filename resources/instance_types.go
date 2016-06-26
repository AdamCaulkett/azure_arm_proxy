package resources

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/rightscale/azure_arm_proxy/config"
)

// SetupGroupsRoutes declares routes for resource group resource
func SetupInstanceTypesRoutes(e *echo.Group) {
	e.Get("/locations/:location/instance_types", listInstanceTypes)
}

//This API lists all available virtual machine sizes for a subscription in a given region.
func listInstanceTypes(c *echo.Context) error {
	location := c.Param("location")
	creds, err := GetClientCredentials(c)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s/subscriptions/%s/providers/Microsoft.Compute/locations/%s/vmSizes?api-version=%s", config.BaseURL, creds.Subscription, location, microsoftComputeApiVersion)
	its, err := GetResources(c, path)
	if err != nil {
		return err
	}

	//TODO: add hrefs or use AzureResource interface
	return c.JSON(200, its)
}
