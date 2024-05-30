# hc
hcはAtCoder Heuristic Contestのローカルテストをサポートするためのcliツールです。
シングルバイナリのシンプルな構成かつインタラクティブなセットアップで導入時の面倒な作業を極力なくし、テスト実行からログの記録までサポートします。


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
[release](https://github.com/ynzwtks/hc/releases)より、最新の実行ファイル(hc)をダウンロードして任意のフォルダに配置してパスを通します。
または下記コマンドでインストールします。

```shell
go install github.com/ynzwtks/hc@latest
```
　
ファイルの構成は以下の通りです。cli本体以外はcli実行時に自動で生成されます。

| 要素  |　ファイル |
| ------------- | ------------- |
|    cli本体 | hc       |
| 定義ファイル   | contest.toml   |
| ログ    |  {BaseDir}/logs/* |
| テストケース定義 | {BaseDir}/test/* |

## How to use
詳細は各コマンドのヘルプを参照してください。

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
  - 公式から提供されるツール(gen、vis or tester)
  - ジャッジ対象のプログラム
  
2. コンテストの評価方法の確認  
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

- テストケース実行時の標準出力結果を確認したい
  -　　runコマンドで「-t」でテストケースを指定した場合は、カレントディレクトリの「out.txt」に結果が出力されます。
  - それ以外はtest/{SetName/配下のxxxx_o.txtに結果が出力されます。 

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
    [{SetName}]
    SetName = '{SetName}'
    TestDataPath = 'test/{SetName}/'
    TestDataNum = 3000
    ExFields = 'A B'  
    IsSystemTest = false
    ```
- テストの実行結果のみ表示させたい
   - runコマンドに「-q」フラグをつけて実行します。(結果は相乗平均、相加平均、テストケース数、エラー件数の順に表示)
      ```shell
      $hc run -q
      10512 20835 30 0
      ```
- cli実行時に特定の環境変数を設定したい
  - 下記の例のようにcontest.tomlのenvセクションにKeysとValuesキーに対となるように定義を追加します。
    ```toml
    [env]
    Keys = ['DEBUG']
    Values = ['true']
    ```
- Google Cloud Run Jobsで並列実行させたい
  - 　今後追記予定
