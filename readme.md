## 概要
AtCoder Heuristic Contestのローカルテストをサポートするためのcliツールです。
シングルバイナリ&インタラクティブなセットアップで導入時の面倒な作業を極力なくし、テスト実行からログの記録まで簡単に実行することができます。

### 主な機能
- ローカルマシンでのテストの並列実行
- テストセットの生成と切り替え
- テスト実行中の経過表示
- テストケースの入力パラーメータをキーにしたフィルタリング実行
- ローカル版公式standingツールとのログの連携
- Cloud Run ジョブによるテスト並列実行
  
### 実行イメージ
- ローカルテストの実行
  <img width="938" alt="スクリーンショット 2024-05-29 19 54 50" src="https://github.com/ynzwtks/hc/assets/73768325/a043364e-7920-47c4-8d31-3f395603ce5b">

- ログの表示
  <img width="543" alt="スクリーンショット 2024-05-29 20 02 25" src="https://github.com/ynzwtks/hc/assets/73768325/cf83f8c4-4ac4-4aab-8891-2d5366b0c3ac">

- Cloud Runジョブによる並列実行
<img width="690" alt="スクリーンショット 2024-05-29 20 18 55" src="https://github.com/ynzwtks/hc/assets/73768325/baa974ff-5287-4c9d-b90c-d0442b24ccb7">

## 動作環境
- macOS ※AHC030,AHC031,AHC033で動作確認済
- Windows(動作未検証)
- Linux(動作未検証)

## Quick Start
### 事前準備
- 実行プログラムの準備
    - Atcoder公式から提供されるツール
        - テスト生成プログラム(gen)
        - スコア計算プログラム(vis or tester)

       Windows以外の場合、下記でビルドを行い
          
         cargo build --release

      ./target/release/配下にファイルが出力されます
            
  - ジャッジ対象のプログラム
  
- コンテスト情報の確認
    - スコアの評価方法(スコアの最大化 or 最小化)
    - インタラクティブ問題か否か
### Install
releasesより最新のバイナリファイルをダウンロードして配置してパスを通します。

https://github.com/ynzwtks/hc/releases

### 環境設定
コンフィグの初期化を行います。
```shell
hc config new
```
対話形式で設定をおこいます。
```shell
hc config setup
```
テストセットを追加します。(set1という名前でseed0からテストケース100件)
```shell
hc config add test -s set1 -c 100
```
システムテストを追加します。(公式でシステムテストのseed値とresult.csvが公開されていることが前提)
```shell
hc config add system -n ahc031
```
テストセットの一覧を表示します。
```shell
hc config list
```
デフォルトのテストセットを切り替えます。
```shell
hc config switch -s {セット名}
```
テストセットのパラメータ毎の分布を確認します。
```shell
hc web param -s {セット名}
```
### テスト実行
テストケースを指定して実行します。
```shell
hc run -t {テスト番号}
```
パラメータ条件を指定してテストセットを実行します。<br>

```shell
ahccli run -t {テスト番号}　-f "N<10 && D<10"
```
テストセットを実行してログに記録します。
```shell
ahccli run -w "test"
```
途中経過は表示せず実行結果のみ表示させて実行します。
```shell
hc run -q
```
### 実行結果の確認
結果のサマリを確認します。
```shell
hc log
```
テストケース毎の結果の確認します。
```shell
hc log　　{ログ番号}
```
テストセットとパラメータ条件を指定してテストケース毎の結果の確認します。
```shell
hc log　　{ログ番号}　 -s {セット名} -f "N < 20"
```
実行結果のテストケース毎の差分をソートして表示します。
```shell
hc log　　diff {ログ番号} {ログ番号} -o asc
```
公式のstandingツールを起動します。
```shell
hc web standings
```

###
INPUT


## Cloud　Run　Jobsを利用した並列実行
今後追記予定

CONTEST_CONFIG_FILE=

フォルダ構成

hc

ファイル・フォルダ構成
hc
{BASE_DIR}/contest.toml  |  コンテスト名
{BASE_DIR}/logs/{テストセット}/
{BASE_DIR}/{テストセット}/

hcはカレントディレクトリ内のcontest.tomlを参照します。

設定ファイル読み込み
・カレントディレクトリ内のcontest.toml
・環境変数「CONTEST_CONFIG_FILE」に定義しているcontest.tomlファイルのパス


デバッグで評価対象プログラム内でどのテストケースが実行されているか知りたい場合は、
環境変数「INPUT_FILE」を参照することで実行中のテストケースのファイル名を参照することができます。
{BASE_DIR}/{テストセット}/


キーを追加する場合はcontest.toml内の
`
[{セット名}]
SetName = '{セット名}'
TestDataPath = 'test/{セット名}/'
TestDataNum = 3000
ExFields = 'E'  
IsSystemTest = false
`

例ExFilelds = 'E E2'"
また、ex.dat
1 2
1 1
2 2
 :

コマンド

###設定ファイルとディレクトリ構造


###環境設定について

###テストの実行について

　
-tコマンドで単体テストケースの実行の場合

　　

###フィルタリング条件について

###ログについて
　　"-w メッセージ"をオプションをつけてテストケース全量を実行した場合のみログに記録します。
　　 単体テストケースの実行やフィルタリング条件での部分実行の場合、ログには記録しません。

###システムテストについて
　　短期コンテストや古いコンテストでは順位ツールのデータがないため、
　　




