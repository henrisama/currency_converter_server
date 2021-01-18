package main

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/henrisama/currency_converter_server/proto"
)

type server struct {
	mu        *sync.RWMutex
	timestamp int64
	rates     map[string]float64
	pb.UnimplementedConverterServer
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
