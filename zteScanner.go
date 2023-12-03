/*
Anoop S
Dec 2023

library to get connected
devices from specific models
of ZTE router.

*/

package zteScanner

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Scanner struct {
	URL      string
	Username string
	Password string
}

func New(baseURL, username, password string) Scanner {
	return Scanner{
		URL:      baseURL,
		Username: username,
		Password: password,
	}
}

func (s Scanner) GetDevices() (Devices, error) {
	var devicesInfo DevicesInfo

	sessionToken, err := getSessionToken(s.URL)
	if err != nil {
		return nil, err
	}

	loginToken, err := getLoginToken(s.URL)
	if err != nil {
		return nil, err
	}

	err = login(s.URL, s.Username, encodePassword(s.Password, loginToken), sessionToken)
	if err != nil {
		return nil, err
	}

	// Hack to fix session expired, call this endpoint
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/?_type=menuView&_tag=localNetStatus&_=%v", s.URL, time.Now().UnixMilli()), nil)

	_, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/?_type=menuData&_tag=wlan_client_stat_lua.lua&_=%v", s.URL, time.Now().UnixMilli()), nil)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error while getting wlan devices : %v", err)
	}

	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&devicesInfo)
	if err != nil {
		return nil, err
	}

	devices, err := parseDevicesInfo(devicesInfo)

	if err != nil {
		return nil, err

	}

	return devices, nil

}

// GetDevices with ping scan
func (s Scanner) GetDevicesAlive() (Devices, error) {

	devices, err := s.GetDevices()
	ch := make(chan Device)
	var wg sync.WaitGroup

	if err != nil {
		return nil, err
	}

	var devicesAlive Devices

	for _, device := range devices {
		wg.Add(1)
		go func(device Device) {
			defer wg.Done()
			if device.IsAlive() {
				ch <- device
			}
		}(device)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for d := range ch {
		devicesAlive = append(devicesAlive, d)
	}

	return devicesAlive, nil

}

// Force Login, GetDevices and Ping Scan them
func (s Scanner) GetDevicesAliveForce() (Devices, error) {

	devices, err := s.GetDevicesForce()
	ch := make(chan Device)
	var wg sync.WaitGroup

	if err != nil {
		return nil, err
	}

	var devicesAlive Devices

	for _, device := range devices {
		wg.Add(1)
		go func(device Device) {
			defer wg.Done()
			if device.IsAlive() {
				ch <- device
			}
		}(device)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for d := range ch {
		devicesAlive = append(devicesAlive, d)
	}

	return devicesAlive, nil

}

// Force Login and GetDevices
func (s Scanner) GetDevicesForce() (Devices, error) {
	var devicesInfo DevicesInfo

	err := forceLogin(s.URL, s.Username, s.Password)
	if err != nil {
		return nil, err

	}

	// Hack to fix session expired, call this endpoint
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/?_type=menuView&_tag=localNetStatus&_=%v", s.URL, time.Now().UnixMilli()), nil)

	_, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequest("GET", fmt.Sprintf("%s/?_type=menuData&_tag=wlan_client_stat_lua.lua&_=%v", s.URL, time.Now().UnixMilli()), nil)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error while getting wlan devices : %v", err)
	}

	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&devicesInfo)
	if err != nil {
		return nil, err
	}

	devices, err := parseDevicesInfo(devicesInfo)

	if err != nil {
		return nil, err

	}
	return devices, nil

}
