package algos

import (
	"testing"

	pb "github.com/FlorinBalint/flo_lb/proto"
)

func TestRRHandler(t *testing.T) {
	tests := []struct {
		name        string
		backends    []*Backend
		initialIdx  int64
		expectedIdx int
	}{
		{
			name:        "Healthy BE is chosen",
			backends:    []*Backend{upAndReadyBackend(t, 0)},
			initialIdx:  -1,
			expectedIdx: 0,
		},
		{
			name:        "Unready BEs are skipped",
			backends:    []*Backend{unreadyBackend(t, 0), unreadyBackend(t, 1), upAndReadyBackend(t, 2)},
			initialIdx:  -1,
			expectedIdx: 2,
		},
		{
			name:        "Wraps back to first",
			backends:    []*Backend{upAndReadyBackend(t, 0), upAndReadyBackend(t, 1)},
			initialIdx:  1,
			expectedIdx: 2, // TODO
		},
		{
			name:        "Wraps back to first after skiping unready",
			backends:    []*Backend{upAndReadyBackend(t, 0), upAndReadyBackend(t, 0), unreadyBackend(t, 1)},
			initialIdx:  1,
			expectedIdx: 3, // TODO
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rr := &RoundRobin{
				backends: test.backends,
				idx:      test.initialIdx,
				beCount:  int64(len(test.backends)),
			}
			rr.Handler(nil)
			if rr.idx != int64(test.expectedIdx) {
				t.Errorf("Handler() want index %v, got %v", test.expectedIdx, rr.idx)
			}
		})
	}
}

func TestRRRegister(t *testing.T) {
	tests := []struct {
		name          string
		backendUrls   []string
		registeredUrl string
		initialIdx    int64
		expectedIdx   int
	}{
		{
			name:          "Register at index",
			backendUrls:   []string{"http://localhost:8081", "http://localhost:8082"},
			registeredUrl: "http://localhost:8083",
			initialIdx:    2,
			expectedIdx:   2,
		},
		{
			name:          "Register after index",
			backendUrls:   []string{"http://localhost:8081", "http://localhost:8082"},
			registeredUrl: "http://localhost:8083",
			initialIdx:    1,
			expectedIdx:   1,
		},
		{
			name:          "Register with modulo index before",
			backendUrls:   []string{"http://localhost:8081", "http://localhost:8082"},
			registeredUrl: "http://localhost:8083",
			initialIdx:    5,
			expectedIdx:   1,
		},
		{
			name:          "Register with modulo index at end, tries new backend next",
			backendUrls:   []string{"http://localhost:8081", "http://localhost:8082"},
			registeredUrl: "http://localhost:8083",
			initialIdx:    4,
			expectedIdx:   2,
		},
		{
			name:          "Register existing backend doesn't do anything",
			backendUrls:   []string{"http://localhost:8081", "http://localhost:8082"},
			registeredUrl: "http://localhost:8081",
			initialIdx:    4,
			expectedIdx:   4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &pb.BackendConfig{
				Type: &pb.BackendConfig_Static{
					Static: &pb.StaticBackends{
						Urls: test.backendUrls,
					},
				},
			}
			rr, err := NewRoundRobin(cfg)
			if err != nil {
				t.Errorf("error creating round robin algorithm %v", t)
			}
			rr.idx = test.initialIdx
			rr.Register(test.registeredUrl)
			if rr.idx != int64(test.expectedIdx) {
				t.Errorf("rr.Register(), want index %v, got %v", test.expectedIdx, rr.idx)
			}
		})
	}
}

func TestRRUnregister(t *testing.T) {
	tests := []struct {
		name         string
		backendUrls  []string
		removedUrl   string
		initialIdx   int64
		expectedIdx  int
		expectedUrls []string
	}{
		{
			name:         "Unregister at index",
			backendUrls:  []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
			removedUrl:   "http://localhost:8083",
			initialIdx:   2,
			expectedIdx:  -1,
			expectedUrls: []string{"http://localhost:8081", "http://localhost:8082"},
		},
		{
			name:         "Unregister after index",
			backendUrls:  []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
			removedUrl:   "http://localhost:8081",
			initialIdx:   1,
			expectedIdx:  0,
			expectedUrls: []string{"http://localhost:8082", "http://localhost:8083"},
		},
		{
			name:         "Unregister before index",
			backendUrls:  []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
			removedUrl:   "http://localhost:8082",
			initialIdx:   2,
			expectedIdx:  1,
			expectedUrls: []string{"http://localhost:8081", "http://localhost:8083"},
		},
		{
			name:         "Unregister index modulo after",
			backendUrls:  []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
			removedUrl:   "http://localhost:8081",
			initialIdx:   4,
			expectedIdx:  0,
			expectedUrls: []string{"http://localhost:8082", "http://localhost:8083"},
		},
		{
			name:         "Unregister index modulo before",
			backendUrls:  []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
			removedUrl:   "http://localhost:8082",
			initialIdx:   5,
			expectedIdx:  1,
			expectedUrls: []string{"http://localhost:8081", "http://localhost:8083"},
		},
		{
			name:         "Unregister index modulo at",
			backendUrls:  []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
			removedUrl:   "http://localhost:8083",
			initialIdx:   5,
			expectedIdx:  -1,
			expectedUrls: []string{"http://localhost:8081", "http://localhost:8082"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &pb.BackendConfig{
				Type: &pb.BackendConfig_Static{
					Static: &pb.StaticBackends{
						Urls: test.backendUrls,
					},
				},
			}
			rr, err := NewRoundRobin(cfg)
			if err != nil {
				t.Errorf("error creating round robin algorithm %v", t)
			}
			rr.idx = test.initialIdx
			rr.Deregister(test.removedUrl)
			if rr.idx != int64(test.expectedIdx) {
				t.Errorf("rr.Unregister(), want index %v, got %v", test.expectedIdx, rr.idx)
			}
			if int(rr.beCount) != len(test.expectedUrls) {
				t.Errorf("rr.Unregister(), want %v backends, got %v", len(test.expectedUrls), rr.beCount)
			}

			for i, wantUrl := range test.expectedUrls {
				if url := rr.backends[i].URL(); wantUrl != url {
					t.Errorf("backend[%v], want %v backend, got %v", i, wantUrl, url)
				}
			}
		})
	}
}
