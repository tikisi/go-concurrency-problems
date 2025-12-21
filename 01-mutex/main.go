package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

type Counter struct {
	mu sync.RWMutex
	v  int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	c.v++
	c.mu.Unlock()
}

func (c *Counter) Value() int {
	defer c.mu.RUnlock()
	c.mu.RLock()
	return c.v
}

func main() {
	var R, W, N int
	in := bufio.NewReader(os.Stdin)
	if _, err := fmt.Fscan(in, &R, &W, &N); err != nil {
		return
	}

	var cnt Counter
	var wg sync.WaitGroup

	wg.Add(W)
	for i := 0; i < W; i++ {
		go func(n int) {
			defer wg.Done()
			for i := 0; i < n; i++ {
				cnt.Inc()
				time.Sleep(1.0)
			}
		}(N)
	}

	wg.Add(R)
	for i := 0; i < R; i++ {
		go func(n int) {
			defer wg.Done()
			for i := 0; i < n; i++ {
				fmt.Println(cnt.Value())
				time.Sleep(1.0)
			}
		}(N)
	}

	wg.Wait()

	fmt.Printf("Final Value: %d\n", cnt.Value())
}
