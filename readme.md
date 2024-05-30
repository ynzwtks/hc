# hc
hcはAtCoder Heuristic Contestのローカルテストをサポートするためのcliツールです。
シングルバイナリ&インタラクティブなセットアップで導入時の面倒な作業を極力なくし、テスト実行からログの記録まで簡単に実行することができます。


### Screencast
![Screencast](https://github.com/ynzwtks/hc/assets/73768325/667d9765-4cfa-4793-9f54-aabfbe0d8649)


## Features
- ローカルマシンでのテスト並列実行
- テストセットの生成と切り替え
- テスト実行中の経過表示
- テストケースの入力パラーメータをキーにしたフィルタリング実行
- ローカル版公式standingツールとのログの連携
- Cloud Runジョブによる並列実行(別途設定必要)
  
## How to Install
```shell
go install github.com/ynzwtks/hc@latest
```
goをinstallしていない場合は、最新の[release](https://github.com/ynzwtks/hc/releases)より、バイナリファイル(hc)を取得して任意のフォルダに配置の上、パスを通します。
Uninstall時は下記のファイル・フォルダを削除してください。
| ファイル・フォルダ  | 説明 |
| ------------- | ------------- |
| hc             | cli本体  |
| contest.toml   | コンテスト定義フィアル(cli実行時に自動生成)  |
| {BaseDir}/logs | ログ格納フォルダ(cli実行時に自動生成)  |
| {BaseDir}/test | テストケース格納フォルダ(cli実行時に自動生成)  |

## How to use
```
Usage:
  hc [command]

Available Commands:
  config      Configure settings
  jobs        Execute cloud run jobs
  log         Display results
  run         Run the test set
  web         Display standings or the histogram of parameters
```
---
## Quick Start
### 1. 事前準備
1. 実行プログラムの準備  
- 公式から提供されるツール  
  gen、vis(or tester)を準備します。Windows以外はソースからビルドする必要ありますが、以下のようにビルドすると「./target/release/」配下に実行ファイルが出力されます。

```
　　　cargo build --release
```            
  - ジャッジ対象のプログラム
  
2. コンテストの評価方法の確認  
   設定で必要になるため、以下は問題文を読んで事前に把握しておきます。
    - スコアの評価方法(スコアの最大化 or 最小化)
    - インタラクティブ問題か否か
---
### 2. 環境設定
コンフィグの初期化を行います。初期化を行うとカレントディレクトリにcontest.tomlが作成されます。
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
---
### 3. テスト実行

テストケースを指定して実行します。テストケース実行時の標準出力結果はカレントディレクトリのout.txtに出力されます。

```shell
hc run -t {テスト番号}
```

パラメータ条件を指定してテストセットを実行します。<br>
```shell
hc run -t {テスト番号}　-f "N<10 && D<10"
```
テストセットを実行してログに記録します。(「-w」オプションがない場合はログに記録しません。)
```shell
hc run -w "test"
```

途中経過は表示せず実行結果のみ表示させます。

```shell
hc run -q
```

### 4. 実行結果の確認

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
---
## Tips

- contest.toml格納フォルダ以外の場所からcliを実行したい
  - cliではカレントディレクトリのcontest.tomlを読み込みますが、環境変数「CONTEST_CONFIG_FILE」を設定することで任意の場所からcliを実行できるようになります。

    ``` sh
    export CONTEST_CONFIG_FILE=/Users/xxxxx/vsshare/ahc/ahc031/contest.toml
    ```
- どのテストケースが実行されているか評価対象プログラム内で確認したい
  - 環境変数「INPUT_FILE」を参照することで評価対象プロラム側から実行中のテストケースのファイル名を参照することができます。

- テストケースをフィルタリングして実行したい
  - テストケースの１行目のパラメータについては、config setupコマンドで設定します。  
    例えばテストケースの１行目が"W D N"を指す場合は空白区切りで以下のように入力します。
    ```
    $hc config setup
           :
    ? Enter the input header's fields(Optional):  W D N
    ```
  - runコマンド、logコマンドで-fオプションを付与することで例えば"-f N<10"のようにフィルタリングすることができます。
    フィルタリングで利用できるオペレーターは[govaluate](https://github.com/Knetic/govaluate)を参照してください。

- フィルタリングを独自のパラメータを追加して行いたい
  - 例えばパラメータA,Bを追加する場合する場合の手順は下記のとおりです。
    1. テストケース毎のパラメータ値をファイルで定義します。複数定義する場合は、空白区切りにします。値はFloat64の範囲内である必要があります。
        ```
        1 2
        1 1
        2 2
         :
        ```
    2. ファイル名を「ex.dat」としテストケースが可能されているディレクトリ(test/{セット名})に配置します。    
    3. contest.toml内のテストセットの定義セクションの「ExFields」にパラメータ名を定義します。
    ```toml
    [{セット名}]
    SetName = '{セット名}'
    TestDataPath = 'test/{セット名}/'
    TestDataNum = 3000
    ExFields = 'A B'  
    IsSystemTest = false
    ```
- テストの実行結果のみ表示させたい
  ```shell
  $hc run -q
  10512 20835 30 0
  ```
- cli実行時に特定の環境変数を設定したい
  - 下記の例のようにcontest.tomlのenvセクションに直接定義します。
    ```toml
    [env]
    Keys = ['DEBUG']
    Values = ['true']
    ```
