package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	chefc "github.com/go-chef/chef"
	"github.com/hashicorp/go-cleanhttp"
)

//Config is struct for ServerUrl and ApiKey
type Config struct {
	ServerURL string
	APIKey    string
	Zone      string
	Oper      string
	Inputs    Inputs
}

//Inputs need to given to powerdns
type Inputs struct {
	Name    string
	Content string
	Type    string
}

//Resource struct defines node attributes and actions
type Resource struct {
	Node   chefc.Node
	Action string
}

//ChefConfig is chef config struct
type ChefConfig struct {
	Provider chefc.Config
	Resource Resource
	client   *chefc.Client
}

//WebHookData type takes any type of variables
type webHookData struct {
	EndpointID      string   `json:"endpointId"`
	EventID         string   `json:"eventId"`
	EventName       string   `json:"eventName"`
	Timestamp       string   `json:"timestamp"`
	ConfigurationID string   `json:"configurationId"`
	UserData        string   `json:"userData"`
	Data            DataType `json:"data"`
}

type DataType struct {
	SCALR_BEHAVIORS                 string
	SCALR_CLOUD_LOCATION            string
	SCALR_CLOUD_LOCATION_ZONE       string
	SCALR_CLOUD_SERVER_ID           string
	SCALR_COST_CENTER_BC            string
	SCALR_COST_CENTER_ID            string
	SCALR_COST_CENTER_NAME          string
	SCALR_ENV_ID                    string
	SCALR_ENV_NAME                  string
	SCALR_EVENT_BEHAVIORS           string
	SCALR_EVENT_CLOUD_LOCATION      string
	SCALR_EVENT_CLOUD_LOCATION_ZONE string
	SCALR_EVENT_CLOUD_SERVER_ID     string
	SCALR_EVENT_COST_CENTER_BC      string
	SCALR_EVENT_COST_CENTER_ID      string
	SCALR_EVENT_COST_CENTER_NAME    string
	SCALR_EVENT_ENV_ID              string
	SCALR_EVENT_ENV_NAME            string
	SCALR_EVENT_EXTERNAL_IP         string
	SCALR_EVENT_FARM_HASH           string
	SCALR_EVENT_FARM_ID             string
	SCALR_EVENT_FARM_NAME           string
	SCALR_EVENT_FARM_OWNER_EMAIL    string
	SCALR_EVENT_FARM_ROLE_ALIAS     string
	SCALR_EVENT_FARM_ROLE_ID        string
	SCALR_EVENT_FARM_TEAM           string
	SCALR_EVENT_IMAGE_ID            string
	SCALR_EVENT_INSTANCE_FARM_INDEX string
	SCALR_EVENT_INSTANCE_INDEX      string
	SCALR_EVENT_INTERNAL_IP         string
	SCALR_EVENT_ISDBMASTER          string
	SCALR_EVENT_NAME                string
	SCALR_EVENT_PROJECT_BC          string
	SCALR_EVENT_PROJECT_ID          string
	SCALR_EVENT_PROJECT_NAME        string
	SCALR_EVENT_ROLE_NAME           string
	SCALR_EVENT_SERVER_HOSTNAME     string
	SCALR_EVENT_SERVER_ID           string
	SCALR_EVENT_SERVER_TYPE         string
	SCALR_EXTERNAL_IP               string
	SCALR_FARM_HASH                 string
	SCALR_FARM_ID                   string
	SCALR_FARM_NAME                 string
	SCALR_FARM_OWNER_EMAIL          string
	SCALR_FARM_ROLE_ALIAS           string
	SCALR_FARM_ROLE_ID              string
	SCALR_FARM_TEAM                 string
	SCALR_IMAGE_ID                  string
	SCALR_INSTANCE_FARM_INDEX       string
	SCALR_INSTANCE_INDEX            string
	SCALR_INTERNAL_IP               string
	SCALR_ISDBMASTER                string
	SCALR_PROJECT_BC                string
	SCALR_PROJECT_ID                string
	SCALR_PROJECT_NAME              string
	SCALR_ROLE_NAME                 string
	SCALR_SERVER_HOSTNAME           string
	SCALR_SERVER_ID                 string
	SCALR_SERVER_TYPE               string
	SERVER_OS                       string
}

func NewRequest(endpoint string, method string, body []byte) error {
	var bodyReader io.Reader
	var Http *http.Client

	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	bodyReader = bytes.NewReader(body)

	req, err := http.NewRequest(method, url.String(), bodyReader)
	req.Header.Add("Content-Type", "application/json")
	Http = cleanhttp.DefaultClient()
	resp, err := Http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error sending request")
	}
	return nil
}

//ScalrWebhookHomePage is used to handle the webrequests
func ScalrWebhookHomePage(w http.ResponseWriter, r *http.Request) {
	config := new(webHookData)
	err := json.NewDecoder(r.Body).Decode(config)
	if err != nil {
		panic(err)
	}
	var IP string
	if config.Data.SCALR_EVENT_EXTERNAL_IP != "" {
		IP = config.Data.SCALR_EVENT_EXTERNAL_IP
	} else {
		IP = config.Data.SCALR_INTERNAL_IP
	}
	hostname := fmt.Sprintf("%s.acoe.com", config.Data.SCALR_EVENT_SERVER_HOSTNAME)
	body := Config{
		ServerURL: "http://192.168.201.211:8081",
		APIKey:    "2fb896a653f4a3a3c15e",
		Zone:      "acoe.com",
		Inputs: Inputs{
			Name:    hostname,
			Content: IP,
			Type:    "A",
		},
	}
	chefbody := &ChefConfig{
		Provider: chefc.Config{
			BaseURL: "https://192.168.200.244/organizations/acoe-dev/",
			Name:    "chefadmin",
			SkipSSL: true,
		},
		Resource: Resource{
			Node: chefc.Node{
				Name: config.Data.SCALR_EVENT_SERVER_HOSTNAME,
			},
		},
	}

	if strings.ToLower(config.EventName) == "hostup" {
		body.Oper = "Create"
		reqBody, _ := json.Marshal(body)
		err := NewRequest("http://192.168.200.238:3001", "POST", reqBody)
		if err != nil {
			panic(err)
		}
	}
	if strings.ToLower(config.EventName) == "hostdown" {
		body.Oper = "Delete"
		reqBody, _ := json.Marshal(body)
		err := NewRequest("http://192.168.200.238:3001", "POST", reqBody)
		if err != nil {
			panic(err)
		}
		chefbody.Resource.Action = "Delete"
		chefreqBody, _ := json.Marshal(chefbody)
		err = NewRequest("http://192.168.200.238:3002", "POST", chefreqBody)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	http.HandleFunc("/", ScalrWebhookHomePage)
	http.ListenAndServe(":3000", nil)
}

