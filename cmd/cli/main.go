package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/anoop142/zteScanner"
)

func main() {
	alive := flag.Bool("a", false, "alive devices")
	username := flag.String("u", "admin", "username")
	password := flag.String("p", "admin", "password")
	forceLogin := flag.Bool("f", false, "force logout")

	flag.Parse()

	var devices zteScanner.Devices
	var err error
	scanner := zteScanner.New("http://192.168.1.1", *username, *password)

	if *alive {
		devices, err = scanner.GetDevicesAlive()
	} else if *forceLogin {
		devices, err = scanner.GetDevicesForce()
	} else {
		devices, err = scanner.GetDevices()
	}
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		fmt.Println("Host name: ", device.Hostname)
		fmt.Println("MAC: ", device.MacAddress)
		fmt.Println("IP: ", device.IPAddress)
		fmt.Printf("========\n\n")
	}
}
