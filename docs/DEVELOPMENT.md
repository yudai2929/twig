# 開発とテスト

## 開発環境

Go のバージョンはルートの `mise.toml` で指定します。Go の外部ライブラリは使っていません。

```sh
mise install
mise exec -- go build -o ./bin/twig ./cmd/twig
```

利用者向けのインストールは [README](../README.md) に記載しています。配布用バイナリには zsh プラグインが埋め込まれており、`twig enable` が `${XDG_DATA_HOME:-$HOME/.local/share}/twig/twig.zsh` に展開します。開発中のビルドを現在の対話 zsh で試す場合は、ビルド後に `eval "$(./bin/twig enable --shell)"` を実行してください。

## テスト

```sh
mise exec -- go test ./...
mise exec -- go test -race ./...
mise exec -- go vet ./...
```

`e2e` テストは macOS の PTY で対話 zsh を起動し、候補の表示、選択、実行、Tab、Space、ファイル名の引用、候補入力後の再表示を確認します。

候補取得だけを確認する場合は次を実行します。

```sh
./bin/twig complete --buffer 'git c' --json
```

## 実装の場所

- `zsh/twig.zsh`: 入力の監視、候補表示、キー操作
- `internal/collector/`: 候補取得、zsh 補完の観測、ヘルプ解析
- `internal/completion/`: 候補データのエンコードとフィルタリング
- `internal/setup/`: zsh 設定への有効化・無効化の反映
- `e2e/`: 対話 zsh を使う結合テスト
