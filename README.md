# Grep Server
検索対象フォルダパス内のファイル内の文字列を検索し、結果を表示します。

***DEMO:***

![Demo](https://image-url.gif)

## Description
ウェブブラウザからの入力で指定ディレクトリ下にあるファイル内の文字列に対して正規表現検索[^1]を行い、結果をhtmlにしてウェブブラウザに表示します。

[^1]: grep (Globally search for the Regular Expression and Print) 検索を行います。正確には、検索に使用するコマンドはgrepの高機能版[ripgrep-all](https://github.com/phiresky/ripgrep-all)を使います。

## Requirement
* [ripgrep-all](https://github.com/phiresky/ripgrep-all)


## Usage

![png](https://github.com/u1and0/grep-server/blob/u1and0-patch-1/Screenshot%20from%202020-05-27%2009-25-04.png)

最初にページにアクセスした画面です。

1. フォルダパスをフルパスで入力します。
2. 検索キーワードをスペース区切りで入力します。検索キーワードには正規表現を使うことができます。
3. 検索階層数(Lv)を1〜5の間から選択します。数字を増やすと検索速度は落ちますがマッチする可能性が上がります。
4. and検索を行うかor検索を行うかをラジオボタンで選択します。
5. [ 検索 ]ボタンをクリックすると検索が始まります。
6. Help: 開発元githubリンク

![png1](https://github.com/u1and0/grep-server/blob/u1and0-patch-1/Screenshot%20from%202020-05-27%2010-12-46.png_)

* 青字のハイライトはマッチした文字列があるファイルです。
* 黒字に黄色背景はマッチした文字列です。行の最初の数字はマッチした行の行数です。


## Features

### オプション

```grep-server -h
-r string
		DB root directory
-s    OS path split windows backslash
-v    show version
-version
		show version
```

### 正規表現の例

table1: 正規表現の例


## 検索オプション
* フォルダパスをフルパスで入力します。
  * ローカルドライブ外のルートパスはデプロイ時に`-r`(root)オプションで指定することができます。
* case sensitiveはsmart caseが有効です。
  * 小文字だけのキーワードに対しては大文字小文字を無視して検索します。
  * 大文字を含んだキーワードに対しては大文字小文字を区別して検索します。
* 検索階層数(Lv)を1〜5の間から選択します。数字を増やすと検索速度は落ちますがマッチする可能性が上がります。
  * 例えばLv: 2を選択したとき、現在ディレクトリから最大2階層下のファイルまでを検索対象ファイルとします。
* and検索を行うかor検索を行うかをラジオボタンで選択します。
  * and検索ではスペースで区切ったキーワードが全て入った行のみを結果として返します。
  * or検索ではスペースで区切ったキーワードのどれかが入った行を結果として返します。

### log

1検索につき1行の検索履歴を/var/log/grep-server.logに記録します。

```
2020/05/27 08:08:41   13files    557matched lines 22.350msec Keyword: [ \d\d3                          ] Path: [ /home/u1and0/Dropbox/Document/統計 ]
2020/05/27 08:08:49   15files    638matched lines 23.324msec Keyword: [ \d\d5                          ] Path: [ /home/u1and0/Dropbox/Document/統計 ]
2020/05/27 08:17:58    0files      0matched lines 14.906msec Keyword: [ import                         ] Path: [ /home/u1and0/Program/go/srg/u1and0 ]
```

table2: 検索ログの例

|   検索ログの内容      |    検索ログの例                     |
|-----------------------|-------------------------------------|
| 検索日時              |    2020/05/27 08:08:41              |
| マッチしたファイル数  |    xxfiles                          |
| マッチした行数        |    xxxmatched lines                 |
| 検索時間              |    xxxmsec                          |
| 検索キーワード        |    Keyword: [    xxxx    ]          |
| 検索パス              |    Path: [    xxx/xxx/xxx   ]       |


## Installation

```
$ git clone https://github.com/u1and0/grep-server
```

```
$ docker pull u1and0/grep-server
```

## Test

```
$ go test
```


## Deploy

```
$ grep-server -r '\\gr.net\path\to\root' -s
```

or use docker container

```
$ docker run -d -v /home/myname:/home/myname u1and0/grep-server
```

ENTRYPOINTに`grep-server`を指定しているので、イメージ名の後はオプションを書き足して下さい。
オプションが不要であれば `$ docker run -d u1and0/grep-server`だけで立ち上げます。
