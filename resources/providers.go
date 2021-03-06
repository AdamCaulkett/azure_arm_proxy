package resources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/labstack/echo"
	"github.com/rightscale/azure_arm_proxy/config"
	eh "github.com/rightscale/azure_arm_proxy/error_handler"
)

type (
	providerResponseParams struct {
		ID                string      `json:"id"`
		Namespace         string      `json:"namespace"`
		RegistrationState string      `json:"registrationState"`
		ResourceTypes     interface{} `json:"resourceTypes"`
		ApplicationID     string      `json:"applicationID,omitempty"`
		Href              string      `json:"href,omitempty"`
	}

	// Provider is base struct for Azure Provider resource
	Provider struct {
		Name           string `json:"name,omitempty"`
		responseParams providerResponseParams
	}
)

const (
	providerAPIVersion = "2015-01-01"
)

// SetupProviderRoutes declares routes for Provider resource
func SetupProviderRoutes(e *echo.Group) {
	e.Get("/providers", listProviders)
	e.Get("/providers/:provider_name", listOneProvider)
	e.Post("/providers/:provider_name/register", registerProvider)
}

func listProviders(c *echo.Context) error {
	return List(c, new(Provider))
}

func listOneProvider(c *echo.Context) error {
	provider := Provider{
		Name: c.Param("provider_name"),
	}
	return Get(c, &provider)
}

// GetRequestParams is a fake function to support AzureResource by Provider
func (p *Provider) GetRequestParams(c *echo.Context) (interface{}, error) { return nil, nil }

// GetResponseParams is accessor function for getting access to responseParams struct
func (p *Provider) GetResponseParams() interface{} {
	return p.responseParams
}

// GetPath returns full path to the sigle provider
func (p *Provider) GetPath(subscription string) string {
	return fmt.Sprintf("%s/subscriptions/%s/providers/%s?api-version=%s", config.BaseURL, subscription, p.Name, providerAPIVersion)
}

// GetCollectionPath returns full path to the collection of providers
func (p *Provider) GetCollectionPath(_ string, subscription string) string {
	return fmt.Sprintf("%s/subscriptions/%s/providers?api-version=%s", config.BaseURL, subscription, providerAPIVersion)
}

// HandleResponse manage raw cloud response
func (p *Provider) HandleResponse(c *echo.Context, body []byte, actionName string) error {
	if err := json.Unmarshal(body, &p.responseParams); err != nil {
		return eh.GenericException(fmt.Sprintf("got bad response from server: %s", string(body)))
	}
	p.responseParams.Href = p.GetHref(p.responseParams.Namespace)
	return nil
}

// GetContentType returns provider content type
func (p *Provider) GetContentType() string {
	return "vnd.rightscale.provider+json"
}

// GetHref returns provider href
func (p *Provider) GetHref(namespace string) string {
	return fmt.Sprintf("providers/%s", namespace)
}

func registerProvider(c *echo.Context) error {
	provider := new(Provider)
	provider.Name = c.Param("provider_name")
	creds, err := GetClientCredentials(c)
	if err != nil {
		return err
	}
	body, err := GetResource(c, provider.GetPath(creds.Subscription))
	if err != nil {
		return err
	}
	provider.HandleResponse(c, body, "")

	if provider.responseParams.RegistrationState == "NotRegistered" {
		config.Logger.Info("Register required:")
		client, err := GetAzureClient(c)
		if err != nil {
			return err
		}
		creds, err := GetClientCredentials(c)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s/subscriptions/%s/providers/%s/register?api-version=%s", config.BaseURL, creds.Subscription, provider.Name, providerAPIVersion)
		config.Logger.Info("Registering Provider ", provider.Name, path)
		resp, err := client.PostForm(path, nil)
		if err != nil {
			return eh.GenericException(fmt.Sprintf("Error has occurred while registering provider: %v", err))
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return eh.GenericException(fmt.Sprintf("failed to load response body: %s", err))
		}

		provider.HandleResponse(c, body, "")
		return Render(c, 200, provider.GetResponseParams(), provider.GetContentType())
	}

	return eh.GenericException(fmt.Sprintf("Provider %s already registered.", provider.Name))
}
