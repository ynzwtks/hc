

# ahccli
ahcのローカルテストを効率的に実行するためのcliツールです。

## 機能
- ローカルテストを並列実行
- コマンドラインからの設定変更、テストデータ定義の追加
- 実行中の経過表示(NGとなったテストケース、無限ループとなったテストケース、大幅に改善・悪化したテストケース)
- 実行結果の相乗平均、過去実行分とのスコア増減率とランク表示、昇順・降順並べ替え
- パラーメータ指定によるフィルタリング実行、拡張パラーメータ定義
- 公式standingツール(ローカル版)との連携と起動
- google cloud run jobsによるテストの並列実行(別途gcpのsetupが必要&課金が発生)

## 特徴
必要なもの
- gen
- vis
- 評価対象プログラム

コンテストにより異なるもの
- スコアの大小
- インタラクティブ問題か否か

## Install
```shell
go install github.com/ynzwtks/ahccli
```

## 環境設定

### 1.コンフィグの初期化を行います。

```shell
hc config new
```
### 2.対話形式で設定をおこいます。
```shell
hc config setup
```
## テストセットの実行

### ・テストセットを実行する



## Quick Start



```shell
hc run
```
##### 実行結果
```text
    AHC030      set1

 100% |██████████████████████████████████████████████████| (100/100) [6s]

Running         [       ]
Done            100/100
Average(log)    42794004 (17.571909)
ok(fail)        100(0)
vsBest(Last)    52(52)

Date               Ave          Log          Data   vsBest      vsLast      Failed 
02/25 12:44:52    42794004   17.571909       100        52        52         0 []
```
#### ・テストセットを実行して結果のログを記録する



```shell
ahccli run -w "コメントです"
```
#### ・別のテストセットを実行する
下記のようにcontest.tomlにセット定義を追加し、cliで追加したセット定義を指定します。
```toml
[FinalSet]
SetName = "FinalSet"
TestDataPath = "test/final/in"
TestDataNum = 3000
```

```shell
ahccli run -s FinalSet
```

#### ・パラメータをフィルタリングして実行する
#### ・パラメータをフィルタリングして繰り返し実行する


```shell
ahccli log diff 0000 0001 -o desc -c 10
```
``` text
0094    [16 2 0.03]        2120205         4984126      235.08%   -2863921
0098    [13 4 0.07]       24320256        39000000      160.36%  -14679744
0051    [13 4 0.06]       17000000        27000000      158.82%  -10000000
0016    [16 4 0.06]       26545864        41000000      154.45%  -14454136
0073    [14 6 0.19]       29000000        39000000      134.48%  -10000000
0043    [16 9 0.10]       72000000        94000000      130.56%  -22000000
0015    [16 11 0.12]     117000000       144000000      123.08%  -27000000
0065    [14 2 0.03]        3598576         4361053      121.19%    -762477
0013    [16 4 0.11]       27628863        33298142      120.52%   -5669279
0035    [10 3 0.17]       10000000        12000000      120.00%   -2000000
```

ahc0001 
n
10^9 sigma pi/n
gen < seeds.txt
vis <input_file> <output_file>

<script type="text/javascript" async src="https://cdnjs.cloudflare.com/ajax/libs/mathjax/2.7.7/MathJax.js?config=TeX-MML-AM_CHTML">
</script>
<script type="text/x-mathjax-config">
 MathJax.Hub.Config({
 tex2jax: {
 inlineMath: [['$', '$'] ],
 displayMath: [ ['$$','$$'], ["\\[","\\]"] ]
 }
 });
</script>
