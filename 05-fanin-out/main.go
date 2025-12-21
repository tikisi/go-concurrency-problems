package main

import (
	"context"
	"fmt"
	"math/rand/v2"
)

func genRand() int {
	return rand.IntN(500000000000)
}

func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	if n%3 == 0 {
		return n == 3
	}

	for i := 5; i <= n/i; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

func IsPrimeFilter(ctx context.Context, input <-chan int) <-chan int {
	ch := make(chan int)

	go func() {
		defer close(ch)

		for v := range input {
			select {
			case <-ctx.Done():
				return
			default:
				if IsPrime(v) {
					ch <- v
				}
			}
		}

	}()

	return ch
}

func main() {
	const PARALLEL_NUM = 8
	ctx := context.Background()

	pipeline :=
		take(ctx,
			fanIn(ctx,
				fanOut(ctx,
					repeatFn(ctx, genRand),
					IsPrimeFilter,
					PARALLEL_NUM,
				)...,
			),
			100000,
		)

	for v := range pipeline {
		fmt.Println(v)
	}
}
