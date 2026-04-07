package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dokicro/homepage-tui/ui"
)

func fetchServices(client *Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return servicesMsg{err: nil}
		}
		groups, err := client.FetchServices()
		return servicesMsg{groups: groups, err: err}
	}
}

func checkSiteMonitor(client *Client, groupName, serviceName string) tea.Cmd {
	return func() tea.Msg {
		result, err := client.FetchSiteMonitor(groupName, serviceName)
		return siteMonitorMsg{
			groupName:   groupName,
			serviceName: serviceName,
			result:      result,
			err:         err,
		}
	}
}

func checkDockerStatus(client *Client, groupName, serviceName, container, server string) tea.Cmd {
	return func() tea.Msg {
		result, err := client.FetchDockerStatus(container, server)
		return dockerStatusMsg{
			groupName:   groupName,
			serviceName: serviceName,
			result:      result,
			err:         err,
		}
	}
}

func fetchResources(client *Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return resourcesMsg{}
		}

		var res ui.Resources
		var errs int

		cpu, err := client.FetchCPU()
		if err == nil {
			res.CPUUsage = cpu.CPU.Usage
		} else {
			errs++
		}

		mem, err := client.FetchMemory()
		if err == nil {
			res.MemTotal = mem.Memory.Total
			res.MemActive = mem.Memory.Active
		} else {
			errs++
		}

		disk, err := client.FetchDisk("/")
		if err == nil {
			res.DiskSize = disk.Drive.Size
			res.DiskUsed = disk.Drive.Used
			res.DiskMount = disk.Drive.Mount
		} else {
			errs++
		}

		net, err := client.FetchNetwork()
		if err == nil {
			if net.Network.RxSec != nil {
				res.NetRxSec = *net.Network.RxSec
			}
			if net.Network.TxSec != nil {
				res.NetTxSec = *net.Network.TxSec
			}
		} else {
			errs++
		}

		if errs == 4 {
			return resourcesMsg{err: fmt.Errorf("all resource fetches failed")}
		}

		return resourcesMsg{resources: res}
	}
}

func fetchHash(client *Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return hashMsg{}
		}
		hash, err := client.FetchHash()
		return hashMsg{hash: hash, err: err}
	}
}

func tickService(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickServiceMsg(t)
	})
}

func tickResource(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickResourceMsg(t)
	})
}

func tickHash(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickHashMsg(t)
	})
}
