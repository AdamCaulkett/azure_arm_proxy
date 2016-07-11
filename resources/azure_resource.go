package resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/rightscale/azure_arm_proxy/config"
	eh "github.com/rightscale/azure_arm_proxy/error_handler"
)

// AzureResource is interface which should support every resource in order to use generic functions List/Get/Create/Delete
type AzureResource interface {
	// GetRequestParams should return params for sending to the cloud.
	// Decodes body params and populate requestParams struct
	GetRequestParams(*echo.Context) (interface{}, error)
	// GetResponseParams should return response params prepared usually by HandleResponse func.
	GetResponseParams() interface{}
	// GetPath should return path to single resource.
	// Builds path from createParams
	GetPath(string) string
	// GetCollectionPath should return path to collection of resource.
	// Input parameter is a parent id (ex: group_name)
	GetCollectionPath(string, string) string
	// HandleResponse could contain varyity of handlers for different actions if needed but the main aim of it
	// is to get raw response (second param), handle it (ex: unmarshal) and modify response params (responseParams) or response header.
	HandleResponse(*echo.Context, []byte, string) error
	// GetContentType should return content type of the resource
	GetContentType() string
	// GetHref should return href of the resource. Input param is a resource id
	GetHref(string) string
}

// Create new resource
func Create(c *echo.Context, r AzureResource) error {
	client, err := GetAzureClient(c)
	if err != nil {
		return err
	}
	requestParams, err := r.GetRequestParams(c)
	if err != nil {
		return err
	}

	by, err := json.Marshal(requestParams)
	if err != nil {
		eh.GenericException(fmt.Sprintf("Error has occurred while marshaling data: %v", err))
	}
	var reader io.Reader
	reader = bytes.NewBufferString(string(by))
	creds, err := GetClientCredentials(c)
	if err != nil {
		return err
	}
	path := r.GetPath(creds.Subscription)
	request, err := http.NewRequest("PUT", path, reader)
	if err != nil {
		return eh.GenericException(fmt.Sprintf("Error has occurred while creating resource: %v", err))
	}
	request.Header.Add("Content-Type", config.MediaType)
	request.Header.Add("Accept", config.MediaType)
	request.Header.Add("User-Agent", config.UserAgent)
	response, err := client.Do(request)
	if err != nil {
		return eh.GenericException(fmt.Sprintf("Error has occurred while creating resource: %v", err))
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return eh.GenericException(fmt.Sprintf("failed to load response body: %s", err))
	}
	if response.StatusCode >= 400 {
		return eh.GenericException(fmt.Sprintf("Error has occurred while creating resource: %s", string(b)))
	}

	//https://msdn.microsoft.com/en-us/library/azure/mt163601.aspx
	if response.Header.Get("Location") != "" {
		array := strings.Split(response.Header.Get("Location"), "/")
		operationId := strings.Split(array[len(array)-1], "?")[0]
		c.Response().Header().Add("OperationId", operationId)
		return c.NoContent(202)
	}

	if err := r.HandleResponse(c, b, "create"); err != nil {
		return err
	}

	return c.NoContent(201)
}

// Delete resource
func Delete(c *echo.Context, r AzureResource) error {
	client, err := GetAzureClient(c)
	if err != nil {
		return err
	}
	creds, err := GetClientCredentials(c)
	if err != nil {
		return err
	}
	path := r.GetPath(creds.Subscription)
	config.Logger.Info("Delete request:", "path", path)

	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return eh.GenericException(fmt.Sprintf("Error has occurred while deleting resource: %v", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return eh.GenericException(fmt.Sprintf("Error has occurred while deleting resource: %v", err))
	}

	if resp.StatusCode >= 400 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return eh.GenericException(fmt.Sprintf("failed to load response body: %s", err))
		}
		return eh.GenericException(fmt.Sprintf("Error has occurred while deleting resource: %v", string(b)))
	}

	//https://msdn.microsoft.com/en-us/library/azure/mt163601.aspx
	if resp.Header.Get("Location") != "" {
		config.Logger.Info("Header:", "Location", resp.Header.Get("Location"))
		array := strings.Split(resp.Header.Get("Location"), "/")
		operationId := strings.Split(array[len(array)-1], "?")[0]
		c.Response().Header().Add("OperationId", operationId)
		return c.NoContent(202)
	}

	return c.NoContent(204)
}

// Get resource
func Get(c *echo.Context, r AzureResource) error {
	creds, err := GetClientCredentials(c)
	if err != nil {
		return err
	}
	path := r.GetPath(creds.Subscription)
	body, err := GetResource(c, path)
	if err != nil {
		return err
	}

	if err := r.HandleResponse(c, body, "get"); err != nil {
		return err
	}
	return Render(c, 200, r.GetResponseParams(), r.GetContentType())
}

// List gets all resources in scope of subscription or in scope of resource group if "group_name" provided
func List(c *echo.Context, r AzureResource) error {
	groupName := c.Param("group_name")
	creds, err := GetClientCredentials(c)
	if err != nil {
		return err
	}
	resourcePath := r.GetCollectionPath(groupName, creds.Subscription)
	resources, err := GetResources(c, resourcePath)
	if err != nil {
		return err
	}
	//add href for each resource
	for _, resource := range resources {
		resource["href"] = r.GetHref(resource["id"].(string))
	}
	return Render(c, 200, resources, r.GetContentType()+";type=collection")
}

// GetResources makes a call to cloud to get all resources
func GetResources(c *echo.Context, path string) ([]map[string]interface{}, error) {
	client, err := GetAzureClient(c)
	if err != nil {
		return nil, err
	}
	config.Logger.Debug("Get Resources request:", "path", path)
	resp, err := client.Get(path)
	defer resp.Body.Close()
	if err != nil {
		return nil, eh.GenericException(fmt.Sprintf("Error has occurred while requesting resources: %v", err))
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, eh.GenericException(fmt.Sprintf("failed to load response body: %s", err))
	}
	if resp.StatusCode >= 400 {
		return nil, eh.GenericException(fmt.Sprintf("Error has occurred while requesting resources: %s", string(b)))
	}

	var m map[string][]map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		if m["value"] != nil {
			// return resources if unmarshaling is success for 'value' key
			// error occurs if value of the hash is not a []map[string]interface{}
			return m["value"], nil
		}
		var array []map[string]interface{}
		if err := json.Unmarshal(b, &array); err == nil {
			// return resources if unmarshaling is success for array struct
			// error occurs if body contains array of resources "[]map[string]interface{}"
			return array, nil
		}
		return nil, eh.GenericException(fmt.Sprintf("got bad response from server: %s", string(b)))
	}
	return m["value"], nil
}

// GetResource sends requests to the clouds to get resource
func GetResource(c *echo.Context, path string) ([]byte, error) {
	client, err := GetAzureClient(c)
	if err != nil {
		return nil, err
	}
	config.Logger.Info("Get Resource request:", "path", path)
	resp, err := client.Get(path)
	defer resp.Body.Close()
	if err != nil {
		return nil, eh.GenericException(fmt.Sprintf("Error has occurred while requesting resource: %v", err))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, eh.GenericException(fmt.Sprintf("failed to load response body: %s", err))
	}
	if resp.StatusCode == 404 {
		return nil, eh.RecordNotFound(c.Param("id"))
	}
	if resp.StatusCode >= 400 {
		return nil, eh.GenericException(fmt.Sprintf("Error has occurred while requesting resource: %s", string(body)))
	}

	return body, nil
}

// Render sends a JSON resource specific content type response with status code.
func Render(c *echo.Context, code int, resources interface{}, contentType string) error {
	c.Response().Header().Set(echo.ContentType, contentType)
	c.Response().WriteHeader(code)
	return json.NewEncoder(c.Response()).Encode(resources)
}
