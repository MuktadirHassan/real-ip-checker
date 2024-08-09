package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/pion/stun"
)

const GET_IP_URL = "https://api.ipify.org?format=json"

func main() {
	resp, err := http.Get(GET_IP_URL)
	if err != nil {
		fmt.Println("unable to get public ip address. make sure you are connected to internet")
	}
	defer resp.Body.Close()
	publicIp := struct {
		Ip string `json:"ip"`
	}{}
	json.NewDecoder(resp.Body).Decode(&publicIp)

	fmt.Println("Your public IP address is: ", publicIp.Ip)

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("unable to get network interfaces")
	}

	var localIPs []string

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("unable to get ip address for interface ", iface.Name)
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

	fmt.Println("Your local IP addresses are: ")
	for index, ip := range localIPs {
		fmt.Println(index+1, " ", ip)
	}

	conn, err := stun.Dial("udp", "stun.l.google.com:19302")
	if err != nil {
		panic(err)
	}

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	if err := conn.Do(message, func(res stun.Event) {
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			panic(err)
		}

		// Print all network related information
		fmt.Println("Your public IP address is: ", xorAddr.IP)
		fmt.Println("Your public port is: ", xorAddr.Port)

	}); err != nil {
		panic(err)
	}

}
