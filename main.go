package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	natpmp "github.com/jackpal/go-nat-pmp"
)

const GET_IP_URL = "https://api.ipify.org?format=json"
const TEST_PORT = 8080

func main() {
	// Step 1: Get Public IP Address
	resp, err := http.Get(GET_IP_URL)
	if err != nil {
		fmt.Println("Unable to get public IP address. Make sure you are connected to the internet.")
		return
	}
	defer resp.Body.Close()

	publicIp := struct {
		Ip string `json:"ip"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&publicIp); err != nil {
		fmt.Println("Error decoding the public IP response.")
		return
	}

	fmt.Println("Your public IP address is:", publicIp.Ip)

	// Step 2: Get Local IP Addresses
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Unable to get network interfaces.")
		return
	}

	var localIPs []string
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Unable to get IP address for interface", iface.Name)
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Ignore loopback addresses
			if ip.IsLoopback() {
				continue
			}

			// Append the IP to the list
			localIPs = append(localIPs, ip.String())
		}
	}

	fmt.Println("Your local IP addresses are:")
	for index, ip := range localIPs {
		fmt.Println(index+1, "-", ip)
	}

	// Step 3: NAT-PMP Port Mapping
	gateway, err := net.ResolveUDPAddr("udp", "192.168.31.1:5351") // Replace with the gateway address
	if err != nil {
		fmt.Println("Unable to resolve gateway address.")
		return
	}

	client := natpmp.NewClient(gateway.IP)
	// Request a port mapping
	response, err := client.AddPortMapping("tcp", TEST_PORT, TEST_PORT, 60) // Map TEST_PORT for 1 minute
	if err != nil {
		fmt.Println("Failed to map port using NAT-PMP. You may be behind a NAT or the gateway does not support NAT-PMP.")
	} else {
		fmt.Println("NAT-PMP port mapping successful!")
		fmt.Printf("Mapped external port %d to internal port %d\n", response.MappedExternalPort, TEST_PORT)
		if publicIp.Ip != "" {
			fmt.Println("You likely have a public IP address.")
		}
	}

}
