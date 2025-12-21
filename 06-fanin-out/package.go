package main

import (
	"context"
	"sync"
)

func orDone[T any](done <-chan struct{}, in <-chan T) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
				case ch <- v:
				}
			}
		}
	}()

	return ch
}

func repeatFn[T any](ctx context.Context, fn func() T) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)
		for {
			select {
			case ch <- fn():
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func take[T any](ctx context.Context, input <-chan T, num int) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)
		for i := 0; i < num; i++ {
			select {
			case v := <-input:
				select {
				case ch <- v:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func fanIn[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	ch := make(chan T)

	var wg sync.WaitGroup
	wg.Add(len(channels))
	for _, inCh := range channels {
		go func() {
			defer wg.Done()
			for v := range orDone(ctx.Done(), inCh) {
				select {
				case ch <- v:
				case <-ctx.Done():
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func fanOut[T any](ctx context.Context, input <-chan T, fn func(context.Context, <-chan T) <-chan T, num int) []<-chan T {
	channels := make([]<-chan T, num)

	for i := 0; i < num; i++ {
		channels[i] = fn(ctx, input)
	}

	return channels
}
