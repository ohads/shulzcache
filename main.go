package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	shulzcache "github.com/ohads/shulzcache/shuzlcache"
)

func fetchDataFromRemote(id int) (string, error) {
	time.Sleep(2 * time.Second)
	log.Printf("running the actual func for id %d\n", id)
	return fmt.Sprintf("Result for ID %d", id), nil
}

func main() {
	cachedFetch := shulzcache.NewCachedFunction(fetchDataFromRemote)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result, err := cachedFetch(100 + i%2)
			if err == nil {
				log.Println("Got:", result)
			}
		}(i)
	}

	wg.Wait()
}
