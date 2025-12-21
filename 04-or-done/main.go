package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// orDone 関数は、入力チャネル 'c' とキャンセルチャネル 'done' を合成し、
// いずれか一方が閉じられたら終了する新しいチャネルを返します。
// これにより、値の消費を簡潔な for range ループで行いつつ、キャンセルに対応できます [1, 3]。
func orDone(done <-chan struct{}, c <-chan interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream) // valStreamが確実に閉じられるようにする
		for {
			select {
			case <-done: // 外部からのキャンセルを検知
				return
			case v, ok := <-c:
				if !ok {
					return // 入力チャネルが閉じられた場合
				}
				// valStreamへの送信もキャンセル可能でなければならない
				select {
				case valStream <- v:
				case <-done:
					// 2回目のループで自動的に閉じるのでreturnする必要はない
				}
			}
		}
	}()
	return valStream
}

// generator 関数は、乱数を無限に生成し続けますが、doneチャネルを監視しています [4]。
func generator(done <-chan struct{}) <-chan interface{} {
	randStream := make(chan interface{})
	go func() {
		defer fmt.Println("generator: Exited.")
		defer close(randStream)

		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		for {
			select {
			case <-done:
				return // キャンセルされたら即座に終了
			case randStream <- r.Intn(100):
				time.Sleep(100 * time.Millisecond) // 生成をシミュレート
			}
		}
	}()
	return randStream
}

func main() {
	// キャンセル用のコンテキストを設定
	ctx, cancel := context.WithCancel(context.Background())
	done := ctx.Done() // context.ContextのDone()メソッドは、キャンセル時に閉じるチャネルを返します [5]。

	// ジェネレーターを起動
	randStream := generator(done)

	// orDoneを使用して、安全で簡潔な消費ループを開始します [2]。
	count := 0
	required := 3

	// orDoneからのストリームを for rangeで回すことで、キャンセルロジックを省略できる [2]。
	for val := range orDone(done, randStream) {
		fmt.Printf("Received: %v\n", val)
		count++

		if count >= required {
			cancel() // 必要な数に達したら外部からキャンセルシグナルを送る
			// for range ループは orDone 関数が終了するまで継続しますが、
			// orDone はすぐに done チャネルのシグナルを受け取り終了するため、安全に抜けられます。
		}
	}

	// すべての並行処理がクリーンアップされるのを待つ（実際には generator の defer 実行を待っている）
	time.Sleep(200 * time.Millisecond)

	fmt.Println("Done.")
}
