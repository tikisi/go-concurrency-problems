# ファンイン・ファンアウトを使った素数判定パイプライン
Go のチャネル・goroutine・context・ジェネリクスを使って、乱数から素数だけを取り出す並行パイプラインを実装しなさい。
次の要件をすべて満たすプログラムを作成すること。

### 全体の概要

* 無限に整数乱数を生成するステージ
* その乱数列から素数だけを通すフィルタステージを、複数並列に動かす fan-out
* 複数のフィルタの出力を 1 つのチャネルにまとめる fan-in
* そこから先頭の一定個数だけ取り出して出力するステージ

というパイプラインを構成し、最終的に「乱数から見つかった素数を 100000 個」標準出力に表示すること。

すべてのステージは共通の `context.Context` によってキャンセルできるようにし、
context がキャンセルされたら速やかに処理を終了すること。

---

## 実装すべき関数

以下の関数群を実装し、それらを組み合わせて `main` 関数でパイプラインを構成しなさい。
汎用的な関数（`repeatFn`, `take`, `fanIn`, `fanOut`）は **ジェネリクス** を使うこと。

### 1. repeatFn

無限ストリームを生成するステージ。

* シグネチャ
    `func repeatFn[T any](ctx context.Context, fn func() T) <-chan T`
* 仕様
    * fnが生成する値を無限に返すchanを返却する


### 2. take
入力ストリームから指定個数だけ値を取り出すステージ。

* シグネチャ
  `func take[T any](ctx context.Context, input <-chan T, num int) <-chan T`

### 3. genRand
乱数生成用の関数。

* シグネチャ
  `func genRand() int`

* 仕様
  * 0 以上 500000000000 未満の整数乱数を生成して返す（`rand.Intn` を用いてよい）。

### 4. IsPrime
整数が素数かどうかを判定する関数。

* シグネチャ
  `func IsPrime(n int) bool`

* 仕様

  * 引数 `n` が素数なら `true`、そうでなければ `false` を返す。
  * 判定方法は自由だが、2 未満は素数ではない、2 と 3 は素数、以降については試し割りでよい。
    （6k±1 最適化を用いてもよい。）

### 5. IsPrimeFilter
素数だけを通すフィルタステージ。

* シグネチャ
  `func IsPrimeFilter(ctx context.Context, input <-chan int) <-chan int`

* 仕様
  * 受信した値が `IsPrime` で素数と判定されたら、その値を `ch` に送信する。

### 6. fanIn
複数のチャネルを 1 つにまとめるステージ（ジェネリクス）。

* シグネチャ
  `func fanIn[T any](ctx context.Context, channels ...<-chan T) <-chan T`

### 7. fanOut
1 つの入力ストリームを複数のフィルタステージに渡すヘルパー（ジェネリクス）。

* シグネチャ
  `func fanOut[T any](ctx context.Context, input <-chan T, fn func(context.Context, <-chan T) <-chan T, num int) []<-chan T`

* 仕様
  * 長さ `num` のスライスを用意し、`num` 回 `fn(ctx, input)` を呼び出して、その返り値チャネルをスライスに格納する。
  * 最終的に `[]<-chan T` を返す。
  * ここでは `fn` として `IsPrimeFilter` を渡すことを想定している（型 `T` は `int`）。

---

## main 関数でのパイプライン構成

`main` 関数では、上記の関数を用いて次のようなパイプラインを構成しなさい。

1. 並列数 `PARALLEL_NUM` を 8 とする。

2. `context.WithCancel(context.Background())` で `ctx` と `cancel` を生成し、
   `defer cancel()` で終了時にキャンセルされるようにする。

3. 無限乱数ストリームを生成する。

   ```
   randStream := repeatFn[int](ctx, genRand)
   ```

4. 乱数ストリームに対して、素数フィルタを `PARALLEL_NUM` 個並列に動かす fan-out を行う。

   ```
   workers := fanOut[int](ctx, randStream, IsPrimeFilter, PARALLEL_NUM)
   ```

5. すべてのフィルタの出力を fan-in でまとめる。

   ```
   merged := fanIn[int](ctx, workers...)
   ```

6. まとめられたストリームから「先頭 100000 個」の値だけを取り出す。

   ```
   pipeline := take[int](ctx, merged, 100000)
   ```

7. `pipeline` を range で読みながら、各素数を 1 行ずつ `fmt.Println` で標準出力に表示する。

8. すべての出力が終わったら `main` を終了すればよい（`defer cancel()` により context がキャンセルされ、残りの goroutine も順次終了できる構造になっていること）。

---

## 実行時の振る舞い

* プログラムは、0 以上 500000000000 未満の乱数から素数を探索し、素数を 100000 個見つけるまで動作する。
* 素数は 1 行に 1 個ずつ、標準出力へ出力される。
* 並列数は 8 であり、素数判定は同時に複数の goroutine によって行われる。
* `context.Context` を利用したキャンセルにより、パイプライン各ステージは終了時に適切に抜けるようになっていること。

この仕様を満たすプログラムを実装しなさい。


## ヒント
```go
// 乱数生成関数
func genRand() int {
	return rand.IntN(500000000000)
}

// 素数判定関数
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
```