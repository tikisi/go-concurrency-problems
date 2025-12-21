# sync.Condを使った待機と通知の実装


## 問題
ある共有キュー（queue）に対し、workerが、キューにデータが追加されるまで待機し、追加されたら処理を開始する仕組みを実装する。

要件
- 共有データとして 整数スライスのキュー を持つこと。
- 3つのワーカー Goroutine が起動されること。
- 各ワーカーは以下の動作を行うこと：
    - queue が空の場合、sync.Cond を使って待機する
    - queue にデータが入ったら処理を開始し、その値を取り出して表示する
- メイン Goroutine は 1 秒待った後、
    - キューに 整数 42 を追加し、Broadcast を使ってすべてのワーカーに通知すること。
- 全ワーカーが値の処理を行ったあと、プログラムを終了すること。
- 必要に応じてSleep等でタイミングを調整してよい

## 入力
なし

## 出力
```
Consumer 0: waiting
Consumer 2: waiting
Consumer 1: waiting
Producer: produce 0
Producer: produce 1
Producer: produce 2
Consumer 0: received 0
Consumer 1: received 1
Consumer 2: received 2
```
