package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	queue := make([]int, 0)
	c := sync.NewCond(&sync.Mutex{})

	// Consumer
	for i := 0; i < 3; i++ {
		go func(no int) {
			c.L.Lock()

			// Spriour WakeUp対策
			for len(queue) == 0 {
				fmt.Printf("Consumer %d: waiting\n", no)
				c.Wait()
			}

			fmt.Printf("Consumer %d: received %d\n", no, queue[0])
			queue = queue[1:]

			c.L.Unlock()
		}(i)
	}

	// 待機させないとProducerがすぐにデータを入力してしまう
	time.Sleep(1 * time.Second)

	// Producer
	for i := 0; i < 3; i++ {
		c.L.Lock()
		queue = append(queue, i)
		fmt.Printf("Producer: produce %d\n", i)
		c.Broadcast() // c.Notifyでも可
		c.L.Unlock()
	}

	time.Sleep(1 * time.Second)
}
