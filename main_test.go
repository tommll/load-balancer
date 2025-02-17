package main

import (
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type MockLoadBalancer struct {
	requestCh chan string
	servers   []*MockServer
	mutex     sync.Mutex
	healthCh  chan bool
}

type MockServer struct {
	name      string
	requestCh chan string
	alive     bool
	mutex     sync.Mutex
}

func (s *MockServer) isAlive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.alive
}

func (s *MockServer) handleRequest(request string) {
	log.Println("Server", s.name, "handleRequest: ", request)
}

func (s *MockServer) listenAndServe() {
	log.Println("Server", s.name, "is running...")
	for {
		// load balance logic
		request := <-s.requestCh
		log.Println("Server", s.name, "received request: ", request)
		s.handleRequest(request)
	}
}

func (lb *MockLoadBalancer) addServer(server *MockServer) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.servers = append(lb.servers, server)
}

func (lb *MockLoadBalancer) nextServer() *MockServer {
	number := generateRandomNumber()
	if isEven(number) {
		return lb.servers[0]
	}
	return lb.servers[1]
}

func (lb *MockLoadBalancer) sendRequest(request string) {
	lb.requestCh <- request
}

func (lb *MockLoadBalancer) listenAndServe() {
	log.Println("Load balancer is running...")
	for {
		// load balance logic
		request := <-lb.requestCh
		log.Println("Load balancer received request: ", request)
		destServer := lb.nextServer()
		log.Println("Load balancer forwarding request to server: ", destServer.name)
		destServer.handleRequest(request)
	}

}

func (lb *MockLoadBalancer) healthCheck() {
	for {
		select {
		case <-lb.healthCh:
			log.Println("Load balancer is shutting down")
			return
		case <-time.After(3 * time.Second):
			log.Println("Load balancer is checking health of servers")

			aliveServers := []*MockServer{}
			lb.mutex.Lock()
			for _, server := range lb.servers {
				if server.isAlive() {
					aliveServers = append(aliveServers, server)
				}
			}
			lb.servers = aliveServers
			log.Println("Load balancer has", len(lb.servers), "alive servers")
			lb.mutex.Unlock()
		}
	}
}

// Method to generate a random number
func generateRandomNumber() int {
	return rand.Intn(100) // Generate a random number between 0 and 99
}

// Method to check if the number is even
func isEven(number int) bool {
	return number%2 == 0
}

func TestHello(t *testing.T) {
	server1 := MockServer{name: "A", requestCh: make(chan string), alive: true}
	server2 := MockServer{name: "B", requestCh: make(chan string), alive: true}

	go server1.listenAndServe()
	go server2.listenAndServe()

	lb := MockLoadBalancer{requestCh: make(chan string), healthCh: make(chan bool)}

	// add processing servers
	lb.addServer(&server1)
	lb.addServer(&server2)

	go lb.healthCheck()

	go lb.listenAndServe()

	time.Sleep(1 * time.Second)

	for i := 0; i < 3; i++ {
		req := "hello"
		lb.sendRequest(req)
		time.Sleep(1 * time.Second)
	}

	// cancel health check
	lb.healthCh <- true
	log.Println("TestHello finished")
}
