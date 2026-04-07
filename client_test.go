package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient(server.URL, AuthConfig{})
	return server, client
}

func TestClient_FetchServices(t *testing.T) {
	groups := []ServiceGroup{
		{
			Name: "Media",
			Services: []Service{
				{Name: "Sonarr", Href: "http://sonarr.local", Icon: "sonarr.png"},
			},
		},
		{
			Name: "Tools",
			Services: []Service{
				{Name: "Portainer", Href: "http://portainer.local", Description: "Docker UI"},
			},
		},
	}

	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/services" {
			t.Errorf("expected path /api/services, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups)
	})
	defer server.Close()

	result, err := client.FetchServices()
	if err != nil {
		t.Fatalf("FetchServices returned error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(result))
	}
	if result[0].Name != "Media" {
		t.Errorf("expected group name Media, got %s", result[0].Name)
	}
	if len(result[0].Services) != 1 {
		t.Fatalf("expected 1 service in Media group, got %d", len(result[0].Services))
	}
	if result[0].Services[0].Name != "Sonarr" {
		t.Errorf("expected service name Sonarr, got %s", result[0].Services[0].Name)
	}
	if result[1].Name != "Tools" {
		t.Errorf("expected group name Tools, got %s", result[1].Name)
	}
}

func TestClient_FetchSiteMonitor(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/siteMonitor" {
			t.Errorf("expected path /api/siteMonitor, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("groupName") != "Media" {
			t.Errorf("expected groupName=Media, got %s", r.URL.Query().Get("groupName"))
		}
		if r.URL.Query().Get("serviceName") != "Sonarr" {
			t.Errorf("expected serviceName=Sonarr, got %s", r.URL.Query().Get("serviceName"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SiteMonitorResult{Status: 200, Latency: 12.5})
	})
	defer server.Close()

	result, err := client.FetchSiteMonitor("Media", "Sonarr")
	if err != nil {
		t.Fatalf("FetchSiteMonitor returned error: %v", err)
	}
	if result.Status != 200 {
		t.Errorf("expected status 200, got %d", result.Status)
	}
	if result.Latency != 12.5 {
		t.Errorf("expected latency 12.5, got %f", result.Latency)
	}
}

func TestClient_FetchDockerStatus(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/docker/status/sonarr/local" {
			t.Errorf("expected path /api/docker/status/sonarr/local, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DockerStatusResult{Status: "running", Health: "healthy"})
	})
	defer server.Close()

	result, err := client.FetchDockerStatus("sonarr", "local")
	if err != nil {
		t.Fatalf("FetchDockerStatus returned error: %v", err)
	}
	if result.Status != "running" {
		t.Errorf("expected status running, got %s", result.Status)
	}
	if result.Health != "healthy" {
		t.Errorf("expected health healthy, got %s", result.Health)
	}
}

func TestClient_FetchCPU(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "cpu" {
			t.Errorf("expected type=cpu, got %s", r.URL.Query().Get("type"))
		}
		w.Header().Set("Content-Type", "application/json")
		resp := CPUResponse{}
		resp.CPU.Usage = 42.5
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.FetchCPU()
	if err != nil {
		t.Fatalf("FetchCPU returned error: %v", err)
	}
	if result.CPU.Usage != 42.5 {
		t.Errorf("expected CPU usage 42.5, got %f", result.CPU.Usage)
	}
}

func TestClient_FetchMemory(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "memory" {
			t.Errorf("expected type=memory, got %s", r.URL.Query().Get("type"))
		}
		w.Header().Set("Content-Type", "application/json")
		resp := MemoryResponse{}
		resp.Memory.Total = 16000000000
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.FetchMemory()
	if err != nil {
		t.Fatalf("FetchMemory returned error: %v", err)
	}
	if result.Memory.Total != 16000000000 {
		t.Errorf("expected total 16000000000, got %d", result.Memory.Total)
	}
}

func TestClient_FetchDisk(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "disk" {
			t.Errorf("expected type=disk, got %s", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("target") != "/" {
			t.Errorf("expected target=/, got %s", r.URL.Query().Get("target"))
		}
		w.Header().Set("Content-Type", "application/json")
		resp := DiskResponse{}
		resp.Drive.Use = 50.0
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.FetchDisk("/")
	if err != nil {
		t.Fatalf("FetchDisk returned error: %v", err)
	}
	if result.Drive.Use != 50.0 {
		t.Errorf("expected disk use 50.0, got %f", result.Drive.Use)
	}
}

func TestClient_FetchNetwork(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "network" {
			t.Errorf("expected type=network, got %s", r.URL.Query().Get("type"))
		}
		w.Header().Set("Content-Type", "application/json")
		rxSec := 125000.0
		resp := NetworkResponse{}
		resp.Network.RxSec = &rxSec
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.FetchNetwork()
	if err != nil {
		t.Fatalf("FetchNetwork returned error: %v", err)
	}
	if result.Network.RxSec == nil {
		t.Fatal("expected RxSec to be non-nil")
	}
	if *result.Network.RxSec != 125000.0 {
		t.Errorf("expected RxSec 125000.0, got %f", *result.Network.RxSec)
	}
	// TxSec should be nil (nullable)
	if result.Network.TxSec != nil {
		t.Errorf("expected TxSec to be nil, got %f", *result.Network.TxSec)
	}
}

func TestClient_AuthHeaders(t *testing.T) {
	auth := AuthConfig{
		Username: "admin",
		Password: "secret",
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("expected basic auth to be set")
		}
		if username != "admin" {
			t.Errorf("expected username admin, got %s", username)
		}
		if password != "secret" {
			t.Errorf("expected password secret, got %s", password)
		}
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("expected X-Custom-Header=custom-value, got %s", r.Header.Get("X-Custom-Header"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ServiceGroup{})
	}))
	defer server.Close()

	client := NewClient(server.URL, auth)
	_, err := client.FetchServices()
	if err != nil {
		t.Fatalf("FetchServices with auth returned error: %v", err)
	}
}

func TestClient_ServerError(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})
	defer server.Close()

	_, err := client.FetchServices()
	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
}
