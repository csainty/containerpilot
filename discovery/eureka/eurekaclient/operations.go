package eurekaclient

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
)

func (c *Client) RegisterInstance(appId string, instanceInfo *InstanceInfo) error {
	values := []string{"apps", appId}
	path := strings.Join(values, "/")
	instance := &Instance{
		Instance: instanceInfo,
	}
	body, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	var resp *RawResponse
	resp, err = c.Post(path, body)
	if err == nil && resp.StatusCode != http.StatusNoContent {
		err = newError(resp.StatusCode, "Incorrect response from registration", 0)
	}

	return err
}

func (c *Client) UnregisterInstance(appId, instanceId string) error {
	values := []string{"apps", appId, instanceId}
	path := strings.Join(values, "/")
	_, err := c.Delete(path)
	return err
}
func (c *Client) GetApplications() (*Applications, error) {
	response, err := c.Get("apps")
	if err != nil {
		return nil, err
	}
	var applications *Applications = new(Applications)
	err = xml.Unmarshal(response.Body, applications)
	return applications, err
}

func (c *Client) GetApplication(appId string) (*Application, error) {
	values := []string{"apps", appId}
	path := strings.Join(values, "/")
	response, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	var application *Application = new(Application)
	err = xml.Unmarshal(response.Body, application)
	return application, err
}
func (c *Client) GetInstance(appId, instanceId string) (*InstanceInfo, error) {
	values := []string{"apps", appId, instanceId}
	path := strings.Join(values, "/")
	response, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	var instance *InstanceInfo = new(InstanceInfo)
	err = xml.Unmarshal(response.Body, instance)
	return instance, err
}
func (c *Client) SendHeartbeat(appId, instanceId string) error {
	values := []string{"apps", appId, instanceId}
	path := strings.Join(values, "/")
	resp, err := c.Put(path, nil)

	if err == nil && resp.StatusCode != http.StatusOK {
		err = newError(resp.StatusCode, "Incorrect response from SendHeartbeat", 0)
	}

	return err
}
