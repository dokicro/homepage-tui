package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dokicro/homepage-tui/ui"
)

func TestIntegration_FullFlow(t *testing.T) {
	// Mock Homepage API
	mux := http.NewServeMux()

	mux.HandleFunc("/api/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("up"))
	})

	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		groups := []ServiceGroup{
			{
				Name: "Media",
				Services: []Service{
					{Name: "Plex", Href: "http://plex:32400", SiteMonitor: "http://plex:32400"},
					{Name: "Sonarr", Href: "http://sonarr:8989", Container: "sonarr", Server: "local"},
				},
			},
			{
				Name: "Infra",
				Services: []Service{
					{Name: "Traefik", Href: "http://traefik:8080", SiteMonitor: "http://traefik:8080"},
				},
			},
		}
		json.NewEncoder(w).Encode(groups)
	})

	mux.HandleFunc("/api/siteMonitor", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SiteMonitorResult{Status: 200, Latency: 15.0})
	})

	mux.HandleFunc("/api/docker/status/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(DockerStatusResult{Status: "running", Health: "healthy"})
	})

	mux.HandleFunc("/api/widgets/resources", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("type") {
		case "cpu":
			json.NewEncoder(w).Encode(CPUResponse{CPU: struct {
				Usage float64 `json:"usage"`
				Load  float64 `json:"load"`
			}{Usage: 25.0, Load: 1.5}})
		case "memory":
			json.NewEncoder(w).Encode(MemoryResponse{Memory: struct {
				Total  int64 `json:"total"`
				Free   int64 `json:"free"`
				Used   int64 `json:"used"`
				Active int64 `json:"active"`
			}{Total: 16000000000, Active: 8000000000}})
		case "disk":
			json.NewEncoder(w).Encode(DiskResponse{Drive: struct {
				Size      int64   `json:"size"`
				Used      int64   `json:"used"`
				Available int64   `json:"available"`
				Use       float64 `json:"use"`
				Mount     string  `json:"mount"`
			}{Size: 500000000000, Used: 250000000000, Available: 250000000000, Use: 50.0, Mount: "/"}})
		case "network":
			rx := 125000.0
			tx := 62500.0
			json.NewEncoder(w).Encode(NetworkResponse{Network: struct {
				RxBytes int64    `json:"rx_bytes"`
				TxBytes int64    `json:"tx_bytes"`
				RxSec   *float64 `json:"rx_sec"`
				TxSec   *float64 `json:"tx_sec"`
			}{RxSec: &rx, TxSec: &tx}})
		}
	})

	mux.HandleFunc("/api/hash", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(HashResponse{Hash: "abc123"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Create client and verify healthcheck
	client := NewClient(server.URL, AuthConfig{})
	if err := client.Healthcheck(); err != nil {
		t.Fatalf("healthcheck failed: %v", err)
	}

	// Create model
	cfg := Config{
		HomepageURL:     server.URL,
		RefreshInterval: 30 * time.Second,
	}
	m := newModel(cfg, client)
	m.width = 100
	m.height = 30

	// Simulate receiving services
	groups, err := client.FetchServices()
	if err != nil {
		t.Fatalf("fetch services failed: %v", err)
	}
	updated, _ := m.Update(servicesMsg{groups: groups})
	m = updated.(model)

	if len(m.entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(m.entries))
	}

	// Simulate site monitor result
	result, _ := client.FetchSiteMonitor("Media", "Plex")
	updated, _ = m.Update(siteMonitorMsg{groupName: "Media", serviceName: "Plex", result: result})
	m = updated.(model)

	if m.entries[0].HTTPStatus != 200 {
		t.Errorf("Plex status: got %d, want 200", m.entries[0].HTTPStatus)
	}

	// Simulate docker status result
	dockerResult, _ := client.FetchDockerStatus("sonarr", "local")
	updated, _ = m.Update(dockerStatusMsg{groupName: "Media", serviceName: "Sonarr", result: dockerResult})
	m = updated.(model)

	if m.entries[1].DockerState != "running" {
		t.Errorf("Sonarr docker: got %q, want %q", m.entries[1].DockerState, "running")
	}

	// Simulate resources
	cpu, _ := client.FetchCPU()
	mem, _ := client.FetchMemory()
	disk, _ := client.FetchDisk("/")
	net, _ := client.FetchNetwork()

	var netRx, netTx float64
	if net.Network.RxSec != nil {
		netRx = *net.Network.RxSec
	}
	if net.Network.TxSec != nil {
		netTx = *net.Network.TxSec
	}

	updated, _ = m.Update(resourcesMsg{resources: ui.Resources{
		CPUUsage:  cpu.CPU.Usage,
		MemTotal:  mem.Memory.Total,
		MemActive: mem.Memory.Active,
		DiskSize:  disk.Drive.Size,
		DiskUsed:  disk.Drive.Used,
		NetRxSec:  netRx,
		NetTxSec:  netTx,
	}})
	m = updated.(model)

	if m.resources.CPUUsage != 25.0 {
		t.Errorf("CPU usage: got %f, want 25.0", m.resources.CPUUsage)
	}

	// Verify View renders without panic
	view := m.View()
	if view == "" {
		t.Error("View returned empty string")
	}

	// Test navigation
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(model)
	if m.cursor != 1 {
		t.Errorf("cursor after j: got %d, want 1", m.cursor)
	}

	// Test quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command")
	}
}
