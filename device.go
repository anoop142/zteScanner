package zteScanner

import (
	"context"
	"encoding/xml"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type Device struct {
	MacAddress string
	IPAddress  string
	Hostname   string
}

type Devices []Device

type DevicesInfo struct {
	XMLName xml.Name `xml:"ajax_response_xml_root"`
	Text    string   `xml:",chardata"`
	WlanID  struct {
		Text     string `xml:",chardata"`
		Instance []struct {
			Text      string   `xml:",chardata"`
			ParaName  []string `xml:"ParaName"`
			ParaValue []string `xml:"ParaValue"`
		} `xml:"Instance"`
	} `xml:"OBJ_WLAN_AD_ID"`
	LanID struct {
		Text     string `xml:",chardata"`
		Instance []struct {
			Text      string   `xml:",chardata"`
			ParaName  []string `xml:"ParaName"`
			ParaValue []string `xml:"ParaValue"`
		} `xml:"Instance"`
	} `xml:"OBJ_WLANAP_ID"`
}

func parseDevicesInfo(devicesInfo DevicesInfo) (Devices, error) {

	var devices Devices
	var device Device
	requiredDataCount := 3

	for _, instance := range devicesInfo.WlanID.Instance {
		extractedDataCount := 0
		for i := range instance.ParaName {
			if extractedDataCount == requiredDataCount {
				devices = append(devices, device)
				break
			}

			if instance.ParaName[i] == "MACAddress" {
				extractedDataCount += 1
				device.MacAddress = instance.ParaValue[i]
			}

			if instance.ParaName[i] == "HostName" {
				extractedDataCount += 1
				device.Hostname = instance.ParaValue[i]
			}

			if instance.ParaName[i] == "IPAddress" {
				extractedDataCount += 1
				device.IPAddress = instance.ParaValue[i]
			}
		}
	}

	return devices, nil
}

// Ping th e device for 1 second
func (d Device) IsAlive() bool {

	pinger, err := probing.NewPinger(d.IPAddress)
	if err != nil {
		return false
	}
	pinger.Count = 1
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = pinger.RunWithContext(ctx)
	if err != nil {
		return false
	}

	return true
}
