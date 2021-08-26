# Grep Server
検索対象フォルダパス内のファイル内の文字列を検索し、結果を表示します。

***DEMO:***

![Demo](https://github.com/u1and0/grep-server/blob/u1and0-patch-2/Peek%202020-06-07%2023-23.gif)

## Description
ウェブブラウザからの入力で指定ディレクトリ下にあるファイル内の文字列に対して正規表現検索を行い、結果をhtmlにしてウェブブラウザに表示します。

grepの高機能版[ripgrep-all](https://github.com/phiresky/ripgrep-all)を検索に使います。


## Requirement
* [ripgrep-all](https://github.com/phiresky/ripgrep-all)
* [ripgrep](https://github.com/BurntSushi/ripgrep)
* [pandoc](https://pandoc.org/)
* [poppler-utils](https://poppler.freedesktop.org/)
* [ffmpeg](https://ffmpeg.org/)


## Usage

![png](https://github.com/u1and0/grep-server/blob/u1and0-patch-1/Screenshot%20from%202020-05-27%2009-25-04.png)

最初にページにアクセスした画面です。

1. フォルダパスをフルパスで入力します。
2. 検索キーワードをスペース区切りで入力します。検索キーワードには正規表現を使うことができます。
3. 検索階層数(Lv)を1〜5の間から選択します。数字を増やすと検索速度は落ちますがマッチする可能性が上がります。
4. and検索を行うかor検索を行うかをラジオボタンで選択します。
5. [ Search ]ボタンをクリックすると検索が始まります。
6. Help: 開発元githubリンク

![png1](https://github.com/u1and0/grep-server/blob/u1and0-patch-1/Screenshot%20from%202020-05-27%2010-12-46.png)

* 青字のハイライトはマッチした文字列があるファイルです。
* 黒字に黄色背景はマッチした文字列です。行の最初の数字はマッチした行の行数です。


## Features

### コマンドオプション

```grep-server -h
Usage of grep-server:
  -E string
    	Set default encoding (default "UTF-8")
  -T string
    	DB trim prefix for directory path
  -debug
    	run as debug mode
  -encoding string
    	Set default encoding (default "UTF-8")
  -p int
    	http.ListenAndServe port number. Default access to http://localhost:8080/ (default 8080)
  -port int
    	http.ListenAndServe port number. Default access to http://localhost:8080/ (default 8080)
  -r string
    	Append root directory path
  -root string
    	Append root directory path
  -s	OS path split windows backslash
  -sep
    	OS path split windows backslash
  -t duration
    	Search method timeout (default 10s)
  -timeout duration
    	Search method timeout (default 10s)
  -trim string
    	DB trim prefix for directory path
  -v	show version
  -version
    	show version
```

### 正規表現の例

table1: 正規表現の例
|<div align='left'> 検索キーワード   |<div align='left'>  マッチする行     |
|---------|----------|
|<div align='left'>   apple␣コップ␣xlsx    |<div align='left'>   同じ行に「apple」と「コップ」と「xlsx」が含まれる, 大文字小文字は無視    |
|<div align='left'>   Apple␣banana |<div align='left'>   「Apple」と「banana」が含まれる, 大文字小文字は区別する    |
|<div align='left'>   .(doc\|xls)   |<div align='left'>   「.doc」または「.xls」が含まれる, 大文字小文字は無視 |
|<div align='left'>   jpe?g|<div align='left'>「jpeg」または「jpg」が含まれる, 大文字小文字は無視  |
|<div align='left'>   SW5␣34[bd] |<div align='left'> 「SW5」と「34b」、または「SW5」と「34d」が含まれる  |
|<div align='left'>   SW5␣34[b-d]  |<div align='left'> 「SW5」と「34b」、または「SW5」と「34c」、または「SW5」と「34d」が含まれる, 大文字小文字は区別する  |
|<div align='left'>   2[6-9]SS  |<div align='left'> 「26SS」「27SS」, 「28SS」, 「29SS」のどれかが含まれる, 大文字小文字は区別する  |



### 検索オプション
* フォルダパスをフルパスで入力します。
  * ローカルドライブ外のルートパスはデプロイ時に`-r`(root)オプションでドライブのプレフィックスを指定することができます。
* case sensitiveはsmart caseが有効です。
  * 小文字だけのキーワードに対しては大文字小文字を無視して検索します。
  * 大文字を含んだキーワードに対しては大文字小文字を区別して検索します。
* 検索モードは、検索キーワードに基づいて検索する対象を選択します。
  * Contentは**ファイルの内容**を検索します。
  * Fileは**ファイル名**を検索します。
* 検索階層数(Lv)を1〜5の間から選択します。数字を増やすと検索速度は落ちますがマッチする可能性が上がります。
  * 例えばLv: 2を選択したとき、指定ディレクトリから最大2階層下のファイルまでを検索対象ファイルとします。
* and検索を行うかor検索を行うかをラジオボタンで選択します。
  * and検索ではスペースで区切ったキーワードが全て入った行のみを結果として返します。
  * or検索ではスペースで区切ったキーワードのどれかが入った行を結果として返します。

### 検索ログ

1検索につき1行の検索履歴を/var/log/grep-server.logに記録します。

```
2020/06/02 21:07:15 23 matches 23 matched lines 1 files contained matches 415 files searched 997 bytes printed 14123624 bytes searched 0.021035 seconds spent searching 0.033567 seconds Keyword: [ 机 カタログ                          ] Path: [ /home/u1and0/Dropbox/Document
```

table2: 検索ログの例

|   検索ログの内容      |    検索ログの例                                   |
|-----------------------|---------------------------------------------------|
| 検索日時              |    2020/06/02 21:07:15                            |
| マッチ数              |    23 matches                                     |
| マッチした行数        |    23 matched lines                               |
| マッチしたファイル数  |    1 files contained matches                      |
| 検索したファイル数    |    415 files searched                             |
| 表示したバイト数      |    997 bytes printed                              |
| 検索したバイト数      |    14123624 bytes searched                        |
| 検索にかかった時間    |    0.021035 seconds spent searching               |
| 全体にかかった時間    |    0.033567 seconds                               |
| 検索キーワード        |    Keyword: [ 机 カタログ                  ]      |
| 検索パス              |    Path: [    /home/u1and0/Dropbox/Document   ]   |


## Installation

```
$ go get github.com/u1and0/grep-server
```

or use docker

```
$ docker pull u1and0/grep-server
```


## Test

```
$ go test
```


## Deploy

```
$ grep-server -r '\\gr.net\path\to\root' -s -E SHIFT-JIS -t 5s
```

* `-r` で `\\gr.net\path\to\root'` をドライブのプレフィックスとし、
* `-s` でパスセパレータをスラッシュ"/"からバックスラッシュ"\\"に変え
* `-E` で最初に表示されるページのエンコーディングをSHIFT-JISにし、
* `-t` で検索タイムアウトを5秒に設定します。


or use docker container

```
$ docker run -d -p 8082:8080 -v /home/myname:/home/myname u1and0/grep-server\
  -r '\\gr.net\path\to\root' -s -E SHIFT-JIS -t 5s
```

ENTRYPOINTに`grep-server`を指定しているので、イメージ名の後はオプションを書き足して下さい。
オプションが不要であれば `$ docker run -d u1and0/grep-server`だけで立ち上げます。
