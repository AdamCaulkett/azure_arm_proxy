package resources

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/labstack/echo"
	"github.com/rightscale/azure_arm_proxy/config"
	eh "github.com/rightscale/azure_arm_proxy/error_handler"
)

const (
	availabilitySetPath       = "providers/Microsoft.Compute/availabilitySets"
	availabilitySetApiVersion = "2016-02-01"
)

type (
	availabilitySetResponseParams struct {
		ID         string                 `json:"id,omitempty"`
		Name       string                 `json:"name,omitempty"`
		Location   string                 `json:"location"`
		Tags       interface{}            `json:"tags,omitempty"`
		Etag       string                 `json:"etag,omitempty"`
		Properties map[string]interface{} `json:"properties,omitempty"`
		Href       string                 `json:"href,omitempty"`
	}

	availabilitySetRequestParams struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	availabilitySetCreateParams struct {
		Name     string `json:"name,omitempty"`
		Location string `json:"location,omitempty"`
		Group    string `json:"group_name,omitempty"`
	}
	// AvailabilitySet is base struct for Azure Availability Set resource to store input create params,
	// request create params and response params gotten from cloud.
	AvailabilitySet struct {
		createParams   availabilitySetCreateParams
		requestParams  availabilitySetRequestParams
		responseParams availabilitySetResponseParams
	}
)

// SetupAvailabilitySetRoutes declares routes for AvailabilitySet resource
func SetupAvailabilitySetRoutes(e *echo.Group) {
	e.Get("/availability_sets", listAllAvailabilitySets)

	//nested routes
	group := e.Group("/resource_groups/:group_name/availability_sets")
	group.Get("", listAvailabilitySets)
	group.Get("/:id", listOneAvailabilitySet)
	group.Post("", createAvailabilitySet)
	group.Delete("/:id", deleteAvailabilitySet)
}

func listAvailabilitySets(c *echo.Context) error {
	return List(c, new(AvailabilitySet))
}

func listAllAvailabilitySets(c *echo.Context) error {
	creds, err := GetClientCredentials(c)
	if err != nil {
		return err
	}
	var sets []map[string]interface{}
	as := new(AvailabilitySet)
	path := fmt.Sprintf("%s/subscriptions/%s/resourceGroups?api-version=%s", config.BaseURL, creds.Subscription, resourceGroupApiVersion)
	resourceGroups, err := GetResources(c, path)
	if err != nil {
		return err
	}
	for _, group := range resourceGroups {
		groupName := group["name"].(string)
		path := fmt.Sprintf("%s/subscriptions/%s/resourceGroups/%s/%s?api-version=%s", config.BaseURL, creds.Subscription, groupName, availabilitySetPath, microsoftComputeApiVersion)
		resp, err := GetResources(c, path)
		if err != nil {
			return err
		}

		//add href for each resource
		for _, resource := range resp {
			resource["href"] = as.GetHref(resource["id"].(string))
		}

		sets = append(sets, resp...)
	}

	return Render(c, 200, sets, as.GetContentType()+";type=collection")
}

func listOneAvailabilitySet(c *echo.Context) error {
	availabilitySet := AvailabilitySet{
		createParams: availabilitySetCreateParams{
			Name:  c.Param("id"),
			Group: c.Param("group_name"),
		},
	}
	return Get(c, &availabilitySet)
}

func createAvailabilitySet(c *echo.Context) error {
	availabilitySet := new(AvailabilitySet)
	return Create(c, availabilitySet)
}

func deleteAvailabilitySet(c *echo.Context) error {
	availabilitySet := AvailabilitySet{
		createParams: availabilitySetCreateParams{
			Name:  c.Param("id"),
			Group: c.Param("group_name"),
		},
	}
	return Delete(c, &availabilitySet)
}

// GetRequestParams prepares parameters for create availability set request to the cloud
func (as *AvailabilitySet) GetRequestParams(c *echo.Context) (interface{}, error) {
	err := c.Get("bodyDecoder").(*json.Decoder).Decode(&as.createParams)
	if err != nil {
		return nil, eh.GenericException(fmt.Sprintf("Error has occurred while decoding params: %v", err))
	}
	as.createParams.Group = c.Param("group_name")
	as.requestParams.Name = as.createParams.Name
	as.requestParams.Location = as.createParams.Location

	return as.requestParams, nil
}

// GetResponseParams is accessor function for getting access to responseParams struct
func (as *AvailabilitySet) GetResponseParams() interface{} {
	return as.responseParams
}

// GetPath returns full path to the sigle availability set
func (as *AvailabilitySet) GetPath(subscription string) string {
	return fmt.Sprintf("%s/subscriptions/%s/resourceGroups/%s/%s/%s?api-version=%s", config.BaseURL, subscription, as.createParams.Group, availabilitySetPath, as.createParams.Name, microsoftComputeApiVersion)
}

// GetCollectionPath returns full path to the collection of availability sets
func (as *AvailabilitySet) GetCollectionPath(groupName string, subscription string) string {
	if groupName == "" {
		return fmt.Sprintf("%s/subscriptions/%s/%s?api-version=%s", config.BaseURL, subscription, availabilitySetPath, microsoftComputeApiVersion)
	}
	return fmt.Sprintf("%s/subscriptions/%s/resourceGroups/%s/%s?api-version=%s", config.BaseURL, subscription, groupName, availabilitySetPath, microsoftComputeApiVersion)
}

// HandleResponse manage raw cloud response
func (as *AvailabilitySet) HandleResponse(c *echo.Context, body []byte, actionName string) error {
	if err := json.Unmarshal(body, &as.responseParams); err != nil {
		return eh.GenericException(fmt.Sprintf("got bad response from server: %s", string(body)))
	}
	href := as.GetHref(as.responseParams.ID)
	if actionName == "create" {
		c.Response().Header().Add("Location", href)
	} else if actionName == "get" {
		as.responseParams.Href = href
	}
	return nil
}

// GetContentType returns availability set content type
func (as *AvailabilitySet) GetContentType() string {
	return "vnd.rightscale.availability_set+json"
}

// GetHref returns availability set href
func (as *AvailabilitySet) GetHref(availabilitySetID string) string {
	array := strings.Split(availabilitySetID, "/")
	return fmt.Sprintf("resource_groups/%s/availability_sets/%s", array[len(array)-5], array[len(array)-1])
}
