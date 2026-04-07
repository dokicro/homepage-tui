package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"homepage-tui/ui"
)

func newTestModel() model {
	cfg := Config{
		HomepageURL:     "http://localhost:3000",
		RefreshInterval: 30 * time.Second,
	}
	return newModel(cfg, nil)
}

func TestModel_Navigation(t *testing.T) {
	m := newTestModel()
	m.entries = []ui.ServiceEntry{
		{GroupName: "Media", Name: "Plex"},
		{GroupName: "Media", Name: "Sonarr"},
		{GroupName: "Infra", Name: "Traefik"},
	}

	// Move down
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = m2.(model)
	if m.cursor != 1 {
		t.Errorf("cursor after j: got %d, want 1", m.cursor)
	}

	// Move down again
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = m2.(model)
	if m.cursor != 2 {
		t.Errorf("cursor after j: got %d, want 2", m.cursor)
	}

	// Move down at bottom — should stay
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = m2.(model)
	if m.cursor != 2 {
		t.Errorf("cursor at bottom: got %d, want 2", m.cursor)
	}

	// Move up
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = m2.(model)
	if m.cursor != 1 {
		t.Errorf("cursor after k: got %d, want 1", m.cursor)
	}
}

func TestModel_WindowResize(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = m2.(model)
	if m.width != 120 || m.height != 40 {
		t.Errorf("got %dx%d, want 120x40", m.width, m.height)
	}
}

func TestModel_ServicesMessage(t *testing.T) {
	m := newTestModel()
	groups := []ServiceGroup{
		{
			Name: "Media",
			Services: []Service{
				{Name: "Plex", Href: "http://plex:32400", SiteMonitor: "http://plex:32400"},
				{Name: "Sonarr", Container: "sonarr", Server: "local"},
			},
		},
	}

	m2, _ := m.Update(servicesMsg{groups: groups})
	m = m2.(model)

	if len(m.groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(m.groups))
	}
	if len(m.entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(m.entries))
	}
	if m.entries[0].Name != "Plex" || !m.entries[0].HasMonitor {
		t.Error("Plex should have HasMonitor=true")
	}
	if m.entries[1].Name != "Sonarr" || !m.entries[1].HasDocker {
		t.Error("Sonarr should have HasDocker=true")
	}
}

func TestModel_SiteMonitorMessage(t *testing.T) {
	m := newTestModel()
	m.entries = []ui.ServiceEntry{
		{GroupName: "Media", Name: "Plex", HasMonitor: true},
	}

	m2, _ := m.Update(siteMonitorMsg{
		groupName:   "Media",
		serviceName: "Plex",
		result:      SiteMonitorResult{Status: 200, Latency: 12.5},
	})
	m = m2.(model)

	if m.entries[0].HTTPStatus != 200 {
		t.Errorf("got status %d, want 200", m.entries[0].HTTPStatus)
	}
	if m.entries[0].Latency != 12.5 {
		t.Errorf("got latency %f, want 12.5", m.entries[0].Latency)
	}
}

func TestModel_DockerStatusMessage(t *testing.T) {
	m := newTestModel()
	m.entries = []ui.ServiceEntry{
		{GroupName: "Media", Name: "Sonarr", HasDocker: true},
	}

	m2, _ := m.Update(dockerStatusMsg{
		groupName:   "Media",
		serviceName: "Sonarr",
		result:      DockerStatusResult{Status: "running", Health: "healthy"},
	})
	m = m2.(model)

	if m.entries[0].DockerState != "running" {
		t.Errorf("got docker state %q, want %q", m.entries[0].DockerState, "running")
	}
}

func TestModel_ResourcesMessage(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(resourcesMsg{
		resources: ui.Resources{
			CPUUsage:  42.5,
			MemTotal:  16000000000,
			MemActive: 8000000000,
		},
	})
	m = m2.(model)

	if m.resources.CPUUsage != 42.5 {
		t.Errorf("got CPU %f, want 42.5", m.resources.CPUUsage)
	}
}

func TestModel_TabFocus(t *testing.T) {
	m := newTestModel()
	if m.focused != 0 {
		t.Errorf("initial focus: got %d, want 0", m.focused)
	}

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = m2.(model)
	if m.focused != 1 {
		t.Errorf("after tab: got %d, want 1", m.focused)
	}

	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = m2.(model)
	if m.focused != 0 {
		t.Errorf("after second tab: got %d, want 0", m.focused)
	}
}

func TestModel_ViewNonEmpty(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24
	m.entries = []ui.ServiceEntry{
		{GroupName: "Media", Name: "Plex", HasMonitor: true, HTTPStatus: 200, Latency: 12},
	}
	m.resources = ui.Resources{CPUUsage: 42.5}
	m.lastRefresh = time.Now()

	v := m.View()
	if v == "" {
		t.Error("View returned empty string")
	}
}
