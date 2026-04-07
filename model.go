package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"homepage-tui/ui"
)

// Message types

type servicesMsg struct {
	groups []ServiceGroup
	err    error
}

type siteMonitorMsg struct {
	groupName   string
	serviceName string
	result      SiteMonitorResult
	err         error
}

type dockerStatusMsg struct {
	groupName   string
	serviceName string
	result      DockerStatusResult
	err         error
}

type resourcesMsg struct {
	resources ui.Resources
	err       error
}

type hashMsg struct {
	hash string
	err  error
}

type tickServiceMsg time.Time
type tickResourceMsg time.Time
type tickHashMsg time.Time
type urlOpenedMsg struct{}

// Stubs — replaced by poller.go in Task 6
func fetchServices(client *Client) tea.Cmd                                          { return nil }
func fetchResources(client *Client) tea.Cmd                                         { return nil }
func fetchHash(client *Client) tea.Cmd                                              { return nil }
func tickService(d time.Duration) tea.Cmd                                           { return nil }
func tickResource(d time.Duration) tea.Cmd                                          { return nil }
func tickHash(d time.Duration) tea.Cmd                                              { return nil }
func checkSiteMonitor(client *Client, groupName, serviceName string) tea.Cmd        { return nil }
func checkDockerStatus(client *Client, groupName, serviceName, container, server string) tea.Cmd {
	return nil
}

// Model

type model struct {
	config Config
	client *Client

	groups    []ServiceGroup
	entries   []ui.ServiceEntry
	resources ui.Resources

	configHash  string
	lastRefresh time.Time
	warning     string

	cursor   int
	focused  int // 0=services, 1=resources
	width    int
	height   int
	showHelp bool
}

func newModel(cfg Config, client *Client) model {
	return model{
		config: cfg,
		client: client,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchServices(m.client),
		fetchResources(m.client),
		tickService(m.config.RefreshInterval),
		tickResource(5*time.Second),
		tickHash(60*time.Second),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case servicesMsg:
		if msg.err != nil {
			m.warning = fmt.Sprintf("fetch error: %v", msg.err)
			return m, nil
		}
		m.groups = msg.groups
		m.entries = buildEntries(msg.groups)
		m.lastRefresh = time.Now()
		m.warning = ""
		if m.cursor >= len(m.entries) && len(m.entries) > 0 {
			m.cursor = len(m.entries) - 1
		}
		// Trigger status checks for all entries
		var cmds []tea.Cmd
		for _, e := range m.entries {
			cmds = append(cmds, statusCmds(m.client, e)...)
		}
		return m, tea.Batch(cmds...)

	case siteMonitorMsg:
		for i := range m.entries {
			if m.entries[i].GroupName == msg.groupName && m.entries[i].Name == msg.serviceName {
				if msg.err != nil {
					m.entries[i].Error = msg.err.Error()
				} else {
					m.entries[i].HTTPStatus = msg.result.Status
					m.entries[i].Latency = msg.result.Latency
					m.entries[i].Error = ""
					m.entries[i].Loading = false
				}
				break
			}
		}
		return m, nil

	case dockerStatusMsg:
		for i := range m.entries {
			if m.entries[i].GroupName == msg.groupName && m.entries[i].Name == msg.serviceName {
				if msg.err != nil {
					m.entries[i].Error = msg.err.Error()
				} else {
					m.entries[i].DockerState = msg.result.Status
					m.entries[i].Error = ""
					m.entries[i].Loading = false
				}
				break
			}
		}
		return m, nil

	case resourcesMsg:
		if msg.err != nil {
			m.warning = fmt.Sprintf("resources error: %v", msg.err)
			return m, nil
		}
		m.resources = msg.resources
		return m, nil

	case hashMsg:
		var cmds []tea.Cmd
		cmds = append(cmds, tickHash(60*time.Second))
		if msg.err == nil && m.configHash != "" && m.configHash != msg.hash {
			cmds = append(cmds, fetchServices(m.client))
		}
		if msg.err == nil {
			m.configHash = msg.hash
		}
		return m, tea.Batch(cmds...)

	case tickServiceMsg:
		var cmds []tea.Cmd
		for _, e := range m.entries {
			cmds = append(cmds, statusCmds(m.client, e)...)
		}
		cmds = append(cmds, tickService(m.config.RefreshInterval))
		return m, tea.Batch(cmds...)

	case tickResourceMsg:
		return m, tea.Batch(
			fetchResources(m.client),
			tickResource(5*time.Second),
		)

	case tickHashMsg:
		return m, fetchHash(m.client)

	case urlOpenedMsg:
		return m, nil
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When help is showing, any key dismisses it
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
		return m, nil

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "tab":
		m.focused = (m.focused + 1) % 2
		return m, nil

	case "enter":
		if m.cursor < len(m.entries) && m.entries[m.cursor].Href != "" {
			return m, openURL(m.entries[m.cursor].Href)
		}
		return m, nil

	case "r":
		var cmds []tea.Cmd
		cmds = append(cmds, fetchServices(m.client))
		cmds = append(cmds, fetchResources(m.client))
		for _, e := range m.entries {
			cmds = append(cmds, statusCmds(m.client, e)...)
		}
		return m, tea.Batch(cmds...)

	case "?":
		m.showHelp = true
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return renderHelp(m.width, m.height)
	}

	header := ui.RenderHeader(m.width, m.lastRefresh, m.warning)
	footer := ui.RenderFooter(m.width)

	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	contentHeight := m.height - headerHeight - footerHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	var content string
	if m.width >= 80 {
		// Side-by-side layout
		resourceWidth := 30
		serviceWidth := m.width - resourceWidth
		services := ui.RenderServices(m.entries, m.cursor, serviceWidth, contentHeight)
		resources := ui.RenderResources(m.resources, resourceWidth, contentHeight)
		content = lipgloss.JoinHorizontal(lipgloss.Top, services, resources)
	} else {
		// Stacked layout
		servicesHeight := contentHeight * 2 / 3
		resourcesHeight := contentHeight - servicesHeight
		services := ui.RenderServices(m.entries, m.cursor, m.width, servicesHeight)
		resources := ui.RenderResources(m.resources, m.width, resourcesHeight)
		content = lipgloss.JoinVertical(lipgloss.Left, services, resources)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// Helper functions

func buildEntries(groups []ServiceGroup) []ui.ServiceEntry {
	var entries []ui.ServiceEntry
	for _, g := range groups {
		for _, s := range g.Services {
			entry := ui.ServiceEntry{
				GroupName:  g.Name,
				Name:       s.Name,
				Href:       s.Href,
				HasMonitor: s.SiteMonitor != "",
				HasDocker:  s.Container != "",
				Loading:    true,
			}
			entries = append(entries, entry)
		}
	}
	return entries
}

func statusCmds(client *Client, e ui.ServiceEntry) []tea.Cmd {
	var cmds []tea.Cmd
	if e.HasMonitor {
		cmds = append(cmds, checkSiteMonitor(client, e.GroupName, e.Name))
	}
	if e.HasDocker {
		cmds = append(cmds, checkDockerStatus(client, e.GroupName, e.Name, e.DockerState, ""))
	}
	return cmds
}

func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		_ = cmd.Start()
		return urlOpenedMsg{}
	}
}

func renderHelp(width, height int) string {
	help := strings.Join([]string{
		"",
		"  Homepage TUI - Keyboard Shortcuts",
		"  ==================================",
		"",
		"  j / Down    Move cursor down",
		"  k / Up      Move cursor up",
		"  Enter       Open service URL",
		"  Tab         Switch focus (services/resources)",
		"  r           Force refresh all",
		"  ?           Toggle this help",
		"  q / Ctrl+C  Quit",
		"",
		"  Press any key to dismiss...",
		"",
	}, "\n")

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(help)
}
