package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// Function to fetch external IP using an external API
func getExternalIP() (string, error) {
	fmt.Println("Fetching your external IP address...")
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

// RunTCPServer starts a simple TCP server on a random port
func runTCPServer(port int, done chan bool) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}
	defer ln.Close()

	fmt.Printf("TCP server successfully started, listening on port %d\n", port)
	done <- true

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

// handleConnection handles incoming TCP connections
func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	fmt.Println("Received data:", string(buf[:n]))
}

// Function to check if a specific port is open on the external IP
func checkPortReachability(ip string, port int) (bool, error) {
	fmt.Printf("Checking if port %d on IP %s is reachable from the internet...\n", port, ip)
	url := "https://portchecker.io/api/v1/query"
	reqBody := map[string]interface{}{
		"host":  ip,
		"ports": []int{port},
	}

	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// Debugging: Print the response body in case of error
	fmt.Println("Response body:", string(body))

	// {"error":false,"host":"ip","check":[{"port":2124,"status":false}],"msg":null}
	var result struct {
		Error bool `json:"error"`
		Check []struct {
			Port   int  `json:"port"`
			Status bool `json:"status"`
		} `json:"check"`

		Msg string `json:"msg"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, err
	}

	if len(result.Check) == 0 {
		return false, fmt.Errorf("no check results found")
	}

	return result.Check[0].Status, nil
}

func main() {
	var debug bool
	var port int
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.IntVar(&port, "port", 0, "Port number to use")
	flag.Parse()

	if debug {
		fmt.Println("Debug mode enabled")
	}

	if port != 0 {
		fmt.Println("Port number:", port)
	} else {
		port = 49198
	}

	// Fetch the external IP address
	externalIP, err := getExternalIP()
	if err != nil {
		fmt.Println("Error fetching external IP:", err)
		return
	}
	fmt.Println("Your external IP address is:", externalIP)

	// Start TCP server on the selected port in a goroutine
	fmt.Printf("Starting a TCP server on a random port (%d)...\n", port)
	done := make(chan bool)
	go runTCPServer(port, done)

	// Wait for the server to start
	<-done

	// Allow some time for the server to be ready
	time.Sleep(2 * time.Second)

	// Check if the port is reachable from the outside
	isOpen, err := checkPortReachability(externalIP, port)
	if err != nil {
		fmt.Println("Error checking port reachability:", err)
		return
	}

	// Determine if the IP is dedicated or shared
	if isOpen {
		fmt.Printf("Success! Port %d on IP %s is reachable. Your ISP has likely provided you with a dedicated public IP.\n", port, externalIP)
	} else {
		fmt.Printf("Port %d on IP %s is not reachable. Your ISP has likely provided you with a shared public IP (NATed).\n", port, externalIP)
	}
}
