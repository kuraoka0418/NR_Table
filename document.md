# 処理の仕様

## メインウィンドウ
以下はメインウィンドウの処理仕様（わかりやすく整理）です。

目的
- テーブルデータ一覧を参照・選択し、選択中のテーブルの全フィールドを視認性の高い表形式で確認できるようにする。

構成・レイアウト
- ウィンドウは左右二分割（HSplit）。
  - 左ペイン：上部に機能ボタン（TwoCompare / DataCompare）を横並びで配置、下部はスクロール可能なテーブル一覧（インデックスリスト）で残り全領域を占める。
    - 実装箇所: [`mygui.NewMw`](mygui/mw.go) / [`mygui.Mw`](mygui/mw.go)
  - 右ペイン：選択したテーブル詳細を「縦スクロールのみ」で表示する領域（横スクロール禁止）。
    - 初期比率 3:7（ユーザーが自由に割付を変更可能）。
- 左上のボタン
  - 「Open TwoCompare」 -> [`mygui.OpenTwoCompare`](mygui/twoCompare.go) を新規ウィンドウで開く。
  - 「Open DataCompare」 -> [`mygui.OpenDataCompare`](mygui/dataCompare.go) を新規ウィンドウで開く。

リスト（左ペイン）
- データソースは [`mydata.TableList`](mydata/data.go)（型: [`mydata.TableData`](mydata/data.go)）。
- テーブル数が多くても操作できるようスクロール可能な一覧にする。
- 選択時に右ペインの内容を差し替える（差し替えは最小限の再描画で行う）。

詳細表示（右ペイン）
- 表示方針
  - 全データを表示する（「先頭5個だけ」は不可）。ただし表示コストを下げるため描画は遅延・必要時のみ行う。
  - 配列や構造体メンバはすべて「表形式」で表示する（縦スクロール、横スクロール禁止）。
  - 大量データ（Vxx テーブルや A2S/S2V、WeldParm など）は折りたたみ式（Accordion）にして初期非展開とし、ユーザーが展開したときにのみ描画する。
    - 実装箇所: [`mygui.buildDetailView`](mygui/mw.go)、遅延描画ヘルパ `createLazyGrid` を利用。
- 表示ルール（重要）
  - WeldParm（`WELDPARM` 相当）: 表示列は "H番号" と "値"。例: H001, 0x1234。
  - CalParm（可変パラメータ係数、`DCCALPARM` 型想定）: 各要素を V1, V2, ... とし、列は a, b, c, min, max の5列で表示する。`CalParm` がフラット配列の場合は要素を 5 個ずつグルーピングして同様に表示する。
  - A2S / S2V / Vxx / CalParmDataTable / Navi arrays: それぞれ表形式で index/value の行で表示。要素数が多い場合はページネーションや折りたたみで遅延読み込みする。
- スクロール
  - 右ペインは縦方向のスクロールのみ（VScroll）。表内部は可能な限り横折り返し（セル内折り返し）で対応し、横スクロールは発生させない。

パフォーマンス対策
- 遅延レンダリング: 折りたたみセクションを最初はプレースホルダ表示とし、ユーザーが「展開」したときにのみ大量ラベルを生成する（createLazyGrid）。
- 大量ウィジェット生成の回避: 大きな配列を一度に Label で生成するのではなく、必要に応じて分割生成または簡易テキスト表現を使う。
- 再描画最小化: 右ペインの内容差し替えは、外側コンテナの Objects を差し替えて Refresh するだけに留める。

イベント・操作
- テーブル選択: 左リストで選択すると、選択インデックスに対応する [`mydata.TableData`](mydata/data.go) を右ペインに表示する。
- 折りたたみ展開: セクション展開ボタンを押すと該当セクションのみ描画を実行。
- 比較ウィンドウ起動: 上部ボタンで別ウィンドウ（`TwoCompare` / `DataCompare`）を起動。メインウィンドウは常駐し、複数ウィンドウの操作が可能。

実装参照
- メインウィンドウ本体: [`mygui/mw.go`](mygui/mw.go) （[`mygui.Mw`](mygui/mw.go)、[`mygui.buildDetailView`](mygui/mw.go)）
- 比較ウィンドウ（プロトタイプ）: [`mygui/twoCompare.go`](mygui/twoCompare.go)、[`mygui/dataCompare.go`](mygui/dataCompare.go)
- データ定義 / テーブルリスト: [`mydata/data.go`](mydata/data.go) （[`mydata.TableList`](mydata/data.go)、[`mydata.TableData`](mydata/data.go)）

注意点
- 表示は視認性を優先するため「全データ表示」を原則とするが、実行時の負荷が高い箇所は遅延レンダリングや折りたたみによりユーザー操作でのみ展開する方針とする。
- 必要なら `mydata` 側で CalParm 型を明確にして表示を簡素化（例: []DCCALPARM）することで UI 側の処理をさらに高速化できる。
