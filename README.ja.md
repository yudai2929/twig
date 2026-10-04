# Twig

[English](README.md)

Twig は、入力中のコマンドに補完候補を表示する macOS 向け zsh プラグインです。候補はターミナル内に表示され、コマンド名、サブコマンド、オプション、ファイル名を補完できます。

![git コマンドを補完する Twig](docs/demo.gif)

[MP4 のデモ動画を見る](docs/demo.mp4)

## インストール

### mise でビルド済みバイナリを使う

macOS、zsh、[mise](https://mise.jdx.dev/) が必要です。[GitHub Releases](https://github.com/yudai2929/twig/releases) のビルド済みバイナリをインストールします。Go は不要です。

```sh
mise use -g github:yudai2929/twig
eval "$(mise exec -- twig enable --shell)"
```

現在の zsh と次に開く zsh の両方で Twig が有効になります。以後は `twig disable` と `twig enable` で切り替えられます。

### Go でビルドする

`go install` を使う場合は Go 1.27.1 以降が必要です。Go のバイナリディレクトリを `PATH` に追加してください。

```sh
go install github.com/yudai2929/twig/cmd/twig@latest
eval "$(twig enable --shell)"
```

## 使い方

`git c`、`go mod t`、`codex ex` などを入力すると、候補が自動で表示されます。
Git リポジトリ内では `git checkout fe` や `git switch fe` から、現在のブランチに合う候補を表示します。
コマンドとキー操作の一覧は `twig help` で確認できます。

- `↓` / `↑`: 候補表示中は候補を移動し、それ以外は履歴をたどる
- `Ctrl-X j` / `Ctrl-X k`: 候補を移動する
- `Enter`: 選択した候補を入力する。候補がない場合はコマンドを実行する
- `Esc`: 候補を閉じる
- `Tab`: zsh の標準補完を使う

候補を入力した後、続きの候補があれば表示されます。候補を使わずにコマンドを実行するときは、`Esc` で閉じてから `Enter` を押してください。
履歴をたどった直後は候補を表示せず、入力を編集するかカーソルを動かすと再び表示します。

候補の先頭のアイコンは種類を示します。`⚙` はコマンド・サブコマンド、`⚑` はフラグ、`📄` はファイル、`📁` はフォルダ、`•` はその他の値です。

特定のコマンドで自動補完を止めるには、`TWIG_DISABLED_COMMANDS="kubectl,ssh"` のように設定できます。Tab の標準補完は使えます。

## 対応範囲

Twig は `PATH` にあるコマンドと、利用できる zsh 補完、CLI が生成する補完、CLI のヘルプから候補を探します。補完定義やヘルプの形式によっては、候補が出ない、または zsh 標準のファイル名候補だけが出る場合があります。現在のシェルだけに定義された補完は自動候補へ反映されないことがあります。

## 開発資料

[開発資料の目次](docs/README.md)を参照してください。

## ライセンス

[MIT](LICENSE)
