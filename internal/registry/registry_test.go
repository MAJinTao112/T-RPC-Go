package registry

import (
	"sync"
	"testing"
	"time"
)

func TestDiscoverDoesNotDeadlockWithWaitingWriter(t *testing.T) {
	registry := &Registry{
		services: map[string]map[string]Instance{
			"UserService": {
				"127.0.0.1:9100": {Addr: "127.0.0.1:9100"},
			},
		},
	}

	const readers = 16
	const iterations = 2_000
	start := make(chan struct{})
	stopWriter := make(chan struct{})
	var writer sync.WaitGroup
	var readerWorkers sync.WaitGroup

	writer.Add(1)
	go func() {
		defer writer.Done()
		<-start
		for {
			select {
			case <-stopWriter:
				return
			default:
			}
			registry.mu.Lock()
			registry.services["Writer"] = map[string]Instance{
				"127.0.0.1:9200": {Addr: "127.0.0.1:9200"},
			}
			registry.mu.Unlock()
		}
	}()

	for reader := 0; reader < readers; reader++ {
		readerWorkers.Add(1)
		go func() {
			defer readerWorkers.Done()
			<-start
			for i := 0; i < iterations; i++ {
				instances, err := registry.Discover("UserService")
				if err != nil {
					t.Errorf("Discover: %v", err)
					return
				}
				if len(instances) != 1 || instances[0].Addr != "127.0.0.1:9100" {
					t.Errorf("unexpected instances: %v", instances)
					return
				}
			}
		}()
	}

	close(start)
	readersDone := make(chan struct{})
	go func() {
		readerWorkers.Wait()
		close(readersDone)
	}()

	select {
	case <-readersDone:
	case <-time.After(5 * time.Second):
		close(stopWriter)
		t.Fatal("Discover deadlocked while a writer was waiting")
	}
	close(stopWriter)
	writer.Wait()
}
