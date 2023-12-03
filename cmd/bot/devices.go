package main

import (
	"fmt"
	"strings"

	"github.com/anoop142/zteScanner"
)

type KnownDevice struct {
	Mac    string
	Name   string
	Ignore bool
}

func buildMsg(devs []string) string {
	var msg strings.Builder
	for i, d := range devs {
		fmt.Fprintf(&msg, "%d. %s\n", i+1, d)
	}
	return msg.String()
}

func getDevsAll(scanner zteScanner.Scanner) (string, error) {
	var devs string
	devices, err := scanner.GetDevicesForce()
	if err != nil {
		return "", err
	}

	for _, d := range devices {
		name := ""
		if d.Hostname == "" {
			name = "-"
		} else {
			name = d.Hostname
		}

		devs += fmt.Sprintf("%s : %s\n", name, d.MacAddress)
	}
	if len(devs) == 0 {
		return "", fmt.Errorf("found none devices")
	}
	return devs, nil

}

func getDevsUsingDB(scanner zteScanner.Scanner, m Models) (string, error) {
	var devNames string
	devices, err := scanner.GetDevicesForce()
	if err != nil {
		return "", err
	}
	for _, d := range devices {
		knownDev, _ := m.KnownDevices.Get(d.MacAddress)
		if knownDev != nil {
			if knownDev.Ignore {
				continue
			}
			// append \n for telegram msg formatting
			devNames += knownDev.Name + "\n"
		} else {
			if d.Hostname != "" {
				devNames += d.Hostname + "\n"
			} else {
				devNames += d.MacAddress + "\n"
			}
		}
	}
	if devNames == "" {
		return "no devices found", nil
	}
	return devNames, nil
}

func getDevsAliveUsingDB(scanner zteScanner.Scanner, m Models) (string, error) {
	var devNames string
	devices, err := scanner.GetDevicesAliveForce()
	if err != nil {
		return "", err
	}
	for _, d := range devices {
		knownDev, _ := m.KnownDevices.Get(d.MacAddress)
		if knownDev != nil {
			if knownDev.Ignore {
				continue
			}
			devNames += knownDev.Name + "\n"
		} else {
			if d.Hostname != "" {
				devNames += d.Hostname + "\n"
			} else {
				devNames += d.MacAddress + "\n"
			}
		}
	}
	if devNames == "" {
		return "no devices found", nil
	}
	return devNames, nil
}

func saveDevice(m Models, mac, name string) error {
	dev := &KnownDevice{
		Name:   name,
		Mac:    mac,
		Ignore: false,
	}
	return m.KnownDevices.Insert(dev)
}

func ignoreDevice(m Models, mac, name string) error {
	dev := &KnownDevice{
		Name:   name,
		Mac:    mac,
		Ignore: true,
	}
	return m.KnownDevices.Insert(dev)
}

func getSavedList(m Models) (string, error) {
	var savedList string
	devs, err := m.KnownDevices.GetKnown()
	if err != nil {
		return "", err
	}
	for _, d := range devs {
		savedList += fmt.Sprintf("%s : %s\n", d.Name, d.Mac)
	}
	if savedList == "" {
		return "", fmt.Errorf("empty saved devices")
	}
	return savedList, nil
}

func getIgnoredList(m Models) (string, error) {
	var ignoredList string
	devs, err := m.KnownDevices.GetIgnored()
	if err != nil {
		return "", err
	}
	for _, d := range devs {
		ignoredList += fmt.Sprintf("%s : %s\n", d.Name, d.Mac)
	}
	if ignoredList == "" {
		return "", fmt.Errorf("empty saved devices")
	}
	return ignoredList, nil
}
