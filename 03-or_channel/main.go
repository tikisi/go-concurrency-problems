package main

import (
	"fmt"
	"time"
)

// 複数のチャネルのうちどれか1つでも閉じたらCloseするようなチャネルをつくる
func or(channels ...<-chan struct{}) <-chan struct{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan struct{})
	go func() {
		// 1回の再帰で処理する量を増やした方が効率は良い
		// 0,1のどれかを受信したらdefer close(orDone)が実行される)
		defer close(orDone)
		select {
		case <-channels[0]:
		case <-channels[1]:
		// orDoneも渡すことで、orDoneのCloseを伝搬させgo routineリークを防いでいる
		// channels[2:]は仮にlen(channels) == 2だったとしても空配列を返す
		case <-or(append(channels[2:], orDone)...):
		}
	}()

	return orDone
}

func main() {
	var channels []chan struct{}
	var roChannels []<-chan struct{}

	for i := 0; i < 3; i++ {
		ch := make(chan struct{})
		channels = append(channels, ch)
		roChannels = append(roChannels, ch)
	}

	done := or(roChannels...)

	go func() {
		time.Sleep(3 * time.Second)
		close(channels[0])
	}()

	fmt.Println(<-done)
}
