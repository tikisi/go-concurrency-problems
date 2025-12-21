package main

import (
	"context"
	"fmt"
)

type Number interface {
	~int | ~int64 | ~float64
}

func generator[T Number](ctx context.Context, numbers ...T) <-chan T {
	output := make(chan T)

	go func() {
		defer close(output)
		for _, v := range numbers {
			select {
			case output <- v:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}

func add[T Number](ctx context.Context, input <-chan T, additive T) <-chan T {
	output := make(chan T)

	go func() {
		defer close(output)
		for v := range input {
			select {
			case output <- v + additive:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}

func multiply[T Number](ctx context.Context, input <-chan T, multiply T) <-chan T {
	output := make(chan T)

	go func() {
		defer close(output)
		for v := range input {
			select {
			case output <- v * multiply:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}

func main() {
	// 入力
	var N int
	fmt.Scanf("%d", &N)
	var a = make([]int, N)
	for i := 0; i < N; i++ {
		fmt.Scanf("%d", &a[i])
	}
	var A, B int
	fmt.Scanf("%d %d", &A, &B)

	ctx := context.Background()
	pipeline := multiply(ctx, add(ctx, generator(ctx, a...), A), B)

	for v := range pipeline {
		fmt.Printf("%d ", v)
	}
}
