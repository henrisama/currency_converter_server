package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	pb "github.com/henrisama/currency_converter_server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Response struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

func getenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("failed to get env var: %s", key)
	}
	return val
}

func main() {
	appID := getenv("APP_ID")
	const base = "USD"

	urlFmt := "https://openexchangerates.org/api/latest.json?app_id=%s&base=%s"
	url := fmt.Sprintf(urlFmt, appID, base)
	rsp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer rsp.Body.Close()
	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", data)

	response := new(Response)
	json.Unmarshal(data, response)

	port := 9090
	addr := fmt.Sprintf("localhost:%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", addr, err)
	}
	log.Printf("Listening on: %s...", addr)

	s := &server{
		mu:        new(sync.RWMutex),
		timestamp: response.Timestamp,
		rates:     response.Rates,
	}
	srv := grpc.NewServer()
	pb.RegisterConverterServer(srv, s)
	reflection.Register(srv)
	if err = srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
