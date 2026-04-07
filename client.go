package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// API response types

type ServiceGroup struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
}

type Service struct {
	Name        string `json:"name"`
	Href        string `json:"href"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Container   string `json:"container"`
	Server      string `json:"server"`
	SiteMonitor string `json:"siteMonitor"`
}

type SiteMonitorResult struct {
	Status  int     `json:"status"`
	Latency float64 `json:"latency"`
	Error   string  `json:"error"`
}

type DockerStatusResult struct {
	Status string `json:"status"`
	Health string `json:"health"`
	Error  string `json:"error"`
}

type CPUResponse struct {
	CPU struct {
		Usage float64 `json:"usage"`
		Load  float64 `json:"load"`
	} `json:"cpu"`
}

type MemoryResponse struct {
	Memory struct {
		Total  int64 `json:"total"`
		Free   int64 `json:"free"`
		Used   int64 `json:"used"`
		Active int64 `json:"active"`
	} `json:"memory"`
}

type DiskResponse struct {
	Drive struct {
		Size      int64   `json:"size"`
		Used      int64   `json:"used"`
		Available int64   `json:"available"`
		Use       float64 `json:"use"`
		Mount     string  `json:"mount"`
	} `json:"drive"`
}

type NetworkResponse struct {
	Network struct {
		RxBytes int64    `json:"rx_bytes"`
		TxBytes int64    `json:"tx_bytes"`
		RxSec   *float64 `json:"rx_sec"`
		TxSec   *float64 `json:"tx_sec"`
	} `json:"network"`
}

type HashResponse struct {
	Hash string `json:"hash"`
}

// Client

type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       AuthConfig
}

func NewClient(baseURL string, auth AuthConfig) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		auth: auth,
	}
}

func (c *Client) doGet(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if c.auth.Username != "" {
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
	}

	for k, v := range c.auth.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) FetchServices() ([]ServiceGroup, error) {
	data, err := c.doGet("/api/services")
	if err != nil {
		return nil, err
	}
	var result []ServiceGroup
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshalling services: %w", err)
	}
	return result, nil
}

func (c *Client) FetchSiteMonitor(groupName, serviceName string) (SiteMonitorResult, error) {
	path := fmt.Sprintf("/api/siteMonitor?groupName=%s&serviceName=%s", groupName, serviceName)
	data, err := c.doGet(path)
	if err != nil {
		return SiteMonitorResult{}, err
	}
	var result SiteMonitorResult
	if err := json.Unmarshal(data, &result); err != nil {
		return SiteMonitorResult{}, fmt.Errorf("unmarshalling site monitor: %w", err)
	}
	return result, nil
}

func (c *Client) FetchDockerStatus(container, server string) (DockerStatusResult, error) {
	path := fmt.Sprintf("/api/docker/status/%s/%s", container, server)
	data, err := c.doGet(path)
	if err != nil {
		return DockerStatusResult{}, err
	}
	var result DockerStatusResult
	if err := json.Unmarshal(data, &result); err != nil {
		return DockerStatusResult{}, fmt.Errorf("unmarshalling docker status: %w", err)
	}
	return result, nil
}

func (c *Client) FetchCPU() (CPUResponse, error) {
	data, err := c.doGet("/api/widgets/resources?type=cpu")
	if err != nil {
		return CPUResponse{}, err
	}
	var result CPUResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return CPUResponse{}, fmt.Errorf("unmarshalling CPU response: %w", err)
	}
	return result, nil
}

func (c *Client) FetchMemory() (MemoryResponse, error) {
	data, err := c.doGet("/api/widgets/resources?type=memory")
	if err != nil {
		return MemoryResponse{}, err
	}
	var result MemoryResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return MemoryResponse{}, fmt.Errorf("unmarshalling memory response: %w", err)
	}
	return result, nil
}

func (c *Client) FetchDisk(target string) (DiskResponse, error) {
	path := fmt.Sprintf("/api/widgets/resources?type=disk&target=%s", target)
	data, err := c.doGet(path)
	if err != nil {
		return DiskResponse{}, err
	}
	var result DiskResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return DiskResponse{}, fmt.Errorf("unmarshalling disk response: %w", err)
	}
	return result, nil
}

func (c *Client) FetchNetwork() (NetworkResponse, error) {
	data, err := c.doGet("/api/widgets/resources?type=network")
	if err != nil {
		return NetworkResponse{}, err
	}
	var result NetworkResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return NetworkResponse{}, fmt.Errorf("unmarshalling network response: %w", err)
	}
	return result, nil
}

func (c *Client) FetchHash() (string, error) {
	data, err := c.doGet("/api/hash")
	if err != nil {
		return "", err
	}
	var result HashResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("unmarshalling hash response: %w", err)
	}
	return result.Hash, nil
}

func (c *Client) Healthcheck() error {
	_, err := c.doGet("/api/healthcheck")
	return err
}
