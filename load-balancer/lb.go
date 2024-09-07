package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/uuid"
)

type Backend struct {
	Host        string
	Port        int
	IsHealthy   bool
	NumRequests int
}

func (b *Backend) String() string {
	return fmt.Sprintf("%s:%d", b.Host, b.Port)
}

type Event struct {
	EventName string
	Data      interface{}
}

type LB struct {
	backends []*Backend
	events   chan Event
	strategy BalacingStrategy
}

type IncomingReq struct {
	srcConn net.Conn
	reqId   string
}

var lb *LB

func InitLB() {
	backends := []*Backend{
		{Host: "localhost", Port: 8081, IsHealthy: true},
		{Host: "localhost", Port: 8082, IsHealthy: true},
		{Host: "localhost", Port: 8083, IsHealthy: true},
		{Host: "localhost", Port: 8084, IsHealthy: true},
	}

	lb = &LB{
		events:   make(chan Event),
		backends: backends,
		strategy: NewRoundRobinBalancingStrategy(backends),
	}
}

func (lb *LB) Run() {
	listener, err := net.Listen("tcp", ":9090")

	if err != nil {
		panic(err)
	}

	defer listener.Close()
	log.Println("Load balancer started on :9090")

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("unable top accept connection: %s", err.Error())
			panic(err)
		}
		// Once connection is accepted proxying it to the backend
		go lb.Proxy(IncomingReq{srcConn: connection, reqId: uuid.NewString()})
	}
}

func (lb *LB) Proxy(req IncomingReq) {
	// Get the next backend from the strategy
	backend := lb.strategy.GetNextBackend(req)
	log.Printf("in-req: %s out-req %s", req.reqId, backend.String())

	//Setup backend connection
	backendConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", backend.Host, backend.Port))
	if err != nil {
		log.Printf("unable to connect to backend: %s", err.Error())

		// send back the err to the source
		req.srcConn.Write([]byte("unable to connect to backend"))
		req.srcConn.Close()
		panic(err)

	}

	backend.NumRequests++
	// Proxying the request
	go io.Copy(backendConn, req.srcConn)
	go io.Copy(req.srcConn, backendConn)
}
