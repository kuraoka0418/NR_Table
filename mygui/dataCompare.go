package mygui

import (
	// fmt is removed in favor of strconv formatting below

	"image/color"
	"reflect"
	"strconv"
	"strings"

	"NR_Table/mydata"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// OpenDataCompare はデータ集計比較ウィンドウを開きます。
// 指示書に従い、指定したデータキーについて全テーブルの値を一覧し、
// 最も多い値をハイライトします（複数あればすべてハイライト）。
func OpenDataCompare(app fyne.App) {
	w := app.NewWindow("DataCompare")
	w.Resize(fyne.NewSize(900, 700))

	// 内部キー一覧（順序を保つ）
	internalKeys := []string{
		"WeldCode.Material",
		"WeldCode.Method",
		"WeldParm",  // インデックスが必要
		"A2S_Short", // インデックスが必要
		"S2V_Short", // インデックスが必要
		"CalParm",   // V番号のインデックスが必要
		"V05_Data", "V06_Data", "V08_Data", "V12_Data",
		"V13_Data", "V15_Data", "V18_Data", "V19_Data", "V20_Data",
		"V32_Data", "V34_Data", "V36_Data",
		"V56_Data", "V59_Data", "V68_Data",
		"V94_Data", "V95_Data", "V57_Data", "V93_Data",
	}

	// 表示用の説明文 -> 内部キー のマップと、説明文リストを作成
	descToKey := map[string]string{}
	keyDescs := make([]string, 0, len(internalKeys))
	for _, k := range internalKeys {
		desc := fieldDescriptions[k]
		if desc == "" {
			desc = k
		}
		keyDescs = append(keyDescs, desc)
		descToKey[desc] = k
	}

	keySelect := widget.NewSelect(keyDescs, nil)
	if len(keyDescs) > 0 {
		keySelect.SetSelected(keyDescs[0])
	}

	// 配列等のキー用のインデックス入力欄
	indexEntry := widget.NewEntry()
	indexEntry.SetPlaceHolder("index (0-based)")
	indexEntry.Disable()

	// 指定可能なインデックス範囲を表示するラベル
	rangeLabel := widget.NewLabel("")
	rangeLabel.Hide()

	// ステータス表示と結果表示領域
	status := widget.NewLabel("")
	resultBox := container.NewVBox()

	// 選択中の説明文に応じて、コントロールの有効/無効と範囲表示を切り替えるヘルパ
	updateControls := func(selectedDesc string) {
		key, ok := descToKey[selectedDesc]
		if !ok {
			indexEntry.Disable()
			rangeLabel.Hide()
			return
		}
		switch key {
		case "WeldParm":
			indexEntry.Enable()
			rangeLabel.SetText("指定可能範囲: 0-503")
			rangeLabel.Show()
		case "A2S_Short", "S2V_Short":
			indexEntry.Enable()
			rangeLabel.SetText("指定可能範囲: 0-255")
			rangeLabel.Show()
		case "CalParm":
			indexEntry.Enable()
			rangeLabel.SetText("指定可能範囲: 0-22 (V1-V23)")
			rangeLabel.Show()
		default:
			// Vxx 系の配列
			if len(key) > 0 && key[0] == 'V' {
				indexEntry.Enable()
				rangeLabel.SetText("指定可能範囲: 0-127")
				rangeLabel.Show()
			} else {
				indexEntry.Disable()
				rangeLabel.Hide()
			}
		}
	}

	keySelect.OnChanged = updateControls
	updateControls(keySelect.Selected)

	// 指定したテーブル・キー・インデックスから表示用文字列を取得するヘルパ
	// CalParm の場合は a,b,c,min,max を横に並べた文字列を返す
	getValue := func(tbl *mydata.TableData, key string, idx int) string {
		v := reflect.ValueOf(*tbl)
		switch key {
		case "WeldCode.Material":
			return strconv.Itoa(int(tbl.WeldCode.Material))
		case "WeldCode.Method":
			return strconv.Itoa(int(tbl.WeldCode.Method))
		case "WeldParm":
			if idx < 0 || idx >= len(tbl.WeldParm.Parm) {
				return "<範囲外>"
			}
			// 0x + 4桁の大文字16進数
			v := uint64(tbl.WeldParm.Parm[idx])
			s := strings.ToUpper(strconv.FormatUint(v, 16))
			// pad to 4
			for len(s) < 4 {
				s = "0" + s
			}
			return "0x" + s
		case "A2S_Short":
			if idx < 0 || idx >= len(tbl.A2S_Short.Speed) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.A2S_Short.Speed[idx]))
		case "S2V_Short":
			if idx < 0 || idx >= len(tbl.S2V_Short.Values) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.S2V_Short.Values[idx]))
		case "CalParm":
			// CalParm の場合、指定した V# の a,b,c,min,max を横並びで取得
			rv := v.FieldByName("CalParm")
			if !rv.IsValid() {
				return "<na>"
			}
			if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
				elem := rv.Index(0)
				if elem.Kind() == reflect.Struct {
					// 構造体配列の場合
					if idx < 0 || idx >= rv.Len() {
						return "<範囲外>"
					}
					el := rv.Index(idx)
					aStr := getFloatFieldAsString(el, "A")
					if aStr == "" {
						aStr = formatValue(el.Field(0))
					}
					bStr := getFloatFieldAsString(el, "B")
					if bStr == "" {
						bStr = formatValue(el.Field(1))
					}
					cStr := getFloatFieldAsString(el, "C")
					if cStr == "" {
						cStr = formatValue(el.Field(2))
					}
					minStr := getFloatFieldAsString(el, "Min")
					if minStr == "" {
						minStr = formatValue(el.Field(3))
					}
					maxStr := getFloatFieldAsString(el, "Max")
					if maxStr == "" {
						maxStr = formatValue(el.Field(4))
					}
					return "a=" + aStr + ", b=" + bStr + ", c=" + cStr + ", min=" + minStr + ", max=" + maxStr
				} else if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
					// フラット配列の場合は5個ずつグループ化して V# に対応させる
					base := idx * 5
					if base < 0 || base+4 >= rv.Len() {
						return "<範囲外>"
					}
					a := strconv.FormatFloat(rv.Index(base).Float(), 'g', -1, 64)
					b := strconv.FormatFloat(rv.Index(base+1).Float(), 'g', -1, 64)
					c := strconv.FormatFloat(rv.Index(base+2).Float(), 'g', -1, 64)
					min := strconv.FormatFloat(rv.Index(base+3).Float(), 'g', -1, 64)
					max := strconv.FormatFloat(rv.Index(base+4).Float(), 'g', -1, 64)
					return "a=" + a + ", b=" + b + ", c=" + c + ", min=" + min + ", max=" + max
				}
			}
			return "<unknown CalParm>"
		case "V05_Data":
			if idx < 0 || idx >= len(tbl.V05_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V05_Data[idx]))
		case "V06_Data":
			if idx < 0 || idx >= len(tbl.V06_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V06_Data[idx]))
		case "V08_Data":
			if idx < 0 || idx >= len(tbl.V08_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V08_Data[idx]))
		case "V12_Data":
			if idx < 0 || idx >= len(tbl.V12_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V12_Data[idx]))
		case "V13_Data":
			if idx < 0 || idx >= len(tbl.V13_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V13_Data[idx]))
		case "V15_Data":
			if idx < 0 || idx >= len(tbl.V15_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V15_Data[idx]))
		case "V18_Data":
			if idx < 0 || idx >= len(tbl.V18_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V18_Data[idx]))
		case "V19_Data":
			if idx < 0 || idx >= len(tbl.V19_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V19_Data[idx]))
		case "V20_Data":
			if idx < 0 || idx >= len(tbl.V20_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V20_Data[idx]))
		case "V32_Data":
			if idx < 0 || idx >= len(tbl.V32_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V32_Data[idx]))
		case "V34_Data":
			if idx < 0 || idx >= len(tbl.V34_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V34_Data[idx]))
		case "V36_Data":
			if idx < 0 || idx >= len(tbl.V36_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V36_Data[idx]))
		case "V56_Data":
			if idx < 0 || idx >= len(tbl.V56_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V56_Data[idx]))
		case "V57_Data":
			if idx < 0 || idx >= len(tbl.V57_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V57_Data[idx]))
		case "V59_Data":
			if idx < 0 || idx >= len(tbl.V59_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V59_Data[idx]))
		case "V68_Data":
			if idx < 0 || idx >= len(tbl.V68_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V68_Data[idx]))
		case "V93_Data":
			if idx < 0 || idx >= len(tbl.V93_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V93_Data[idx]))
		case "V94_Data":
			if idx < 0 || idx >= len(tbl.V94_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V94_Data[idx]))
		case "V95_Data":
			if idx < 0 || idx >= len(tbl.V95_Data) {
				return "<範囲外>"
			}
			return strconv.Itoa(int(tbl.V95_Data[idx]))
		default:
			return "<na>"
		}
	}

	// 比較ボタンのアクション（結果表示）
	compareBtn := widget.NewButton("Compare", func() {
		if len(mydata.TableList) == 0 {
			status.SetText("No tables loaded")
			return
		}
		selectedDesc := keySelect.Selected
		if selectedDesc == "" {
			status.SetText("Select a key")
			return
		}
		key, ok := descToKey[selectedDesc]
		if !ok {
			status.SetText("Invalid key")
			return
		}

		idx := -1
		needsIndex := false
		switch key {
		case "WeldParm", "A2S_Short", "S2V_Short", "CalParm":
			needsIndex = true
		default:
			if len(key) > 0 && key[0] == 'V' {
				needsIndex = true
			}
		}

		if needsIndex {
			if indexEntry.Text == "" {
				status.SetText("Index を指定してください")
				return
			}
			if n, err := strconv.Atoi(indexEntry.Text); err == nil {
				idx = n
			} else {
				status.SetText("Invalid index")
				return
			}
		}

		// 各テーブルについて値を取得し、頻度を数える
		counts := map[string]int{}
		values := make([]string, len(mydata.TableList))
		hasError := false
		for i := range mydata.TableList {
			val := getValue(&mydata.TableList[i], key, idx)
			values[i] = val
			if val == "<範囲外>" {
				hasError = true
			}
			counts[val]++
		}

		if hasError {
			status.SetText("エラー: 指定された index が範囲外です")
			return
		}

		// 最も多い出現回数を見つけ、それを持つ値を収集する
		max := 0
		for _, c := range counts {
			if c > max {
				max = c
			}
		}
		maxVals := map[string]struct{}{}
		for v, c := range counts {
			if c == max {
				maxVals[v] = struct{}{}
			}
		}

		// 結果表示領域を作成: テーブルごとに行を表示（Table | Value）
		resultBox.Objects = nil

		grid := container.NewGridWithColumns(2)
		// ヘッダ
		grid.Add(widget.NewLabelWithStyle("Table", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		grid.Add(widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		for i, val := range values {
			grid.Add(widget.NewLabel(strconv.Itoa(i)))
			// 通常テキストはテーマの前景色を使用してダーク/ライトに対応
			col := theme.Color(theme.ColorNameForeground)
			if _, ok := maxVals[val]; ok {
				col = color.RGBA{R: 0, G: 128, B: 0, A: 255} // 最頻値は緑でハイライト
			}
			grid.Add(canvas.NewText(val, col))
		}

		// 結果表示領域を縦スクロールにして十分な高さを確保
		tableScroll := container.NewVScroll(grid)
		tableScroll.SetMinSize(fyne.NewSize(800, 520))
		resultBox.Add(tableScroll)

		resultBox.Refresh()
		status.SetText("Compared " + strconv.Itoa(len(values)) + " tables")
	})

	controls := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Key:"), keySelect,
			widget.NewLabel("Index:"), indexEntry,
			compareBtn, status,
		),
		rangeLabel,
	)

	content := container.NewBorder(controls, nil, nil, nil, container.NewVScroll(resultBox))
	w.SetContent(content)
	w.Show()
}

// formatReflectFloatField は struct 要素から名前で float を取得して文字列化します。
// f が struct で、fieldName が無ければ fallbackIndex を使う。
func formatReflectFloatField(f reflect.Value, fieldName string, fallbackIndex int) string {
	if !f.IsValid() {
		return "<invalid>"
	}
	if f.Kind() != reflect.Struct {
		return "<na>"
	}
	ff := f.FieldByName(fieldName)
	if ff.IsValid() && (ff.Kind() == reflect.Float32 || ff.Kind() == reflect.Float64) {
		return strconv.FormatFloat(ff.Convert(reflect.TypeOf(float64(0))).Float(), 'g', 6, 64)
	}
	// fallback by index
	if fallbackIndex >= 0 && fallbackIndex < f.NumField() {
		fld := f.Field(fallbackIndex)
		if fld.Kind() == reflect.Float32 || fld.Kind() == reflect.Float64 {
			return strconv.FormatFloat(fld.Convert(reflect.TypeOf(float64(0))).Float(), 'g', 6, 64)
		}
	}
	return "<na>"
}
