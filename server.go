package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	pb "github.com/henrisama/currency_converter_server/proto"
)

type server struct {
	client    *http.Client
	appID     string
	mu        *sync.RWMutex
	timestamp int64
	rates     map[string]float64
	pb.UnimplementedConverterServer
}

func newServer(appID string, client *http.Client) *server {
	if client == nil {
		client = http.DefaultClient
	}
	return &server{
		client: client,
		appID:  appID,
		mu:     new(sync.RWMutex),
		rates:  make(map[string]float64),
	}
}

func (s *server) fetchRates(ctx context.Context) error {
	const base = "USD"
	urlFmt := "https://openexchangerates.org/api/latest.json?app_id=%s&base=%s"
	url := fmt.Sprintf(urlFmt, s.appID, base)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	rsp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	response := new(Response)
	json.Unmarshal(data, response)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timestamp = response.Timestamp
	s.rates = response.Rates
	return nil
}

func (s *server) getPair(c1, c2 string) (float64, float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v1, ok := s.rates[c1]
	if !ok {
		return 0, 0, fmt.Errorf("failed to get %s", c1)
	}
	v2, ok := s.rates[c2]
	if !ok {
		return 0, 0, fmt.Errorf("failed to get %s", c2)
	}
	return v1, v2, nil
}

func (s *server) getTimestamp() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.timestamp
}

func (s *server) Convert(ctx context.Context, req *pb.ConvertRequest) (
	*pb.ConvertResponse, error) {

	v1, v2, err := s.getPair(req.From, req.To)
	if err != nil {
		return nil, err
	}
	if v1 == 0 {
		return nil, fmt.Errorf("invalid currency value")
	}
	val := v2 / v1
	return &pb.ConvertResponse{
		Timestamp: s.getTimestamp(),
		FromName:  req.From,
		ToName:    req.To,
		Value:     val,
	}, nil
}
