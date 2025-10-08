package mygui

import (
	"fmt"
	"image/color"
	"reflect"
	"strconv"

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
		"WeldParm",  // requires index
		"A2S_Short", // requires index
		"S2V_Short", // requires index
		"CalParm",   // requires V index
		"V05_Data", "V06_Data", "V08_Data", "V12_Data",
		"V13_Data", "V15_Data", "V18_Data", "V19_Data", "V20_Data",
		"V32_Data", "V34_Data", "V36_Data",
		"V56_Data", "V59_Data", "V68_Data",
		"V94_Data", "V95_Data", "V57_Data", "V93_Data",
	}

	// 説明 -> 内部キー マップと説明リストを作成
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

	// index entry for array-like keys
	indexEntry := widget.NewEntry()
	indexEntry.SetPlaceHolder("index (0-based)")
	indexEntry.Disable()

	// 指定可能範囲を表示するラベル
	rangeLabel := widget.NewLabel("")
	rangeLabel.Hide()

	// status and result area
	status := widget.NewLabel("")
	resultBox := container.NewVBox()

	// helper to enable/disable controls based on selected description and show valid range
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
			// Vxx arrays
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

	// value extractor: returns string representation for given table/internalKey/index
	// CalParm の場合は a,b,c,min,max を横に並べた文字列を返す
	getValue := func(tbl *mydata.TableData, key string, idx int) string {
		v := reflect.ValueOf(*tbl)
		switch key {
		case "WeldCode.Material":
			return fmt.Sprintf("%d", tbl.WeldCode.Material)
		case "WeldCode.Method":
			return fmt.Sprintf("%d", tbl.WeldCode.Method)
		case "WeldParm":
			if idx < 0 || idx >= len(tbl.WeldParm.Parm) {
				return "<範囲外>"
			}
			return fmt.Sprintf("0x%04X", tbl.WeldParm.Parm[idx])
		case "A2S_Short":
			if idx < 0 || idx >= len(tbl.A2S_Short.Speed) {
				return "<範囲外>"
			}
			return fmt.Sprintf("%d", tbl.A2S_Short.Speed[idx])
		case "S2V_Short":
			if idx < 0 || idx >= len(tbl.S2V_Short.Values) {
				return "<範囲外>"
			}
			return fmt.Sprintf("%d", tbl.S2V_Short.Values[idx])
		case "CalParm":
			// CalParm の場合、指定された V# の a,b,c,min,max を横並びで表示
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
					return fmt.Sprintf("a=%s, b=%s, c=%s, min=%s, max=%s", aStr, bStr, cStr, minStr, maxStr)
				} else if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
					// フラット配列の場合、5個ずつグループ化
					base := idx * 5
					if base < 0 || base+4 >= rv.Len() {
						return "<範囲外>"
					}
					a := fmt.Sprintf("%.6g", rv.Index(base).Float())
					b := fmt.Sprintf("%.6g", rv.Index(base+1).Float())
					c := fmt.Sprintf("%.6g", rv.Index(base+2).Float())
					min := fmt.Sprintf("%.6g", rv.Index(base+3).Float())
					max := fmt.Sprintf("%.6g", rv.Index(base+4).Float())
					return fmt.Sprintf("a=%s, b=%s, c=%s, min=%s, max=%s", a, b, c, min, max)
				}
			}
			return "<unknown CalParm>"
		default:
			// try to map Vxx fields by reflection
			f := v.FieldByName(key)
			if f.IsValid() && (f.Kind() == reflect.Array || f.Kind() == reflect.Slice) {
				if idx < 0 || idx >= f.Len() {
					return "<範囲外>"
				}
				return formatValue(f.Index(idx))
			}
			// fallback: try direct field
			f2 := v.FieldByName(key)
			if f2.IsValid() {
				return formatValue(f2)
			}
			return "<na>"
		}
	}

	// compare action
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

		// find max frequency
		max := 0
		for _, c := range counts {
			if c > max {
				max = c
			}
		}
		// collect all values with max freq
		maxVals := map[string]struct{}{}
		for v, c := range counts {
			if c == max {
				maxVals[v] = struct{}{}
			}
		}

		// build result UI: per-table table (2 columns: Table | Value)
		resultBox.Objects = nil

		// create a grid with 2 columns: Table index, Value
		grid := container.NewGridWithColumns(2)
		// header
		grid.Add(widget.NewLabelWithStyle("Table", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		grid.Add(widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		for i, val := range values {
			grid.Add(widget.NewLabel(fmt.Sprintf("%d", i)))
			// 通常テキストはテーマの foreground を使い、ダーク/ライト両対応にする
			col := theme.ForegroundColor()
			if _, ok := maxVals[val]; ok {
				col = color.RGBA{R: 0, G: 128, B: 0, A: 255}
			}
			grid.Add(canvas.NewText(val, col))
		}

		// make the table area taller so it occupies most of the window vertically
		tableScroll := container.NewVScroll(grid)
		tableScroll.SetMinSize(fyne.NewSize(800, 520))
		resultBox.Add(tableScroll)

		resultBox.Refresh()
		status.SetText(fmt.Sprintf("Compared %d tables", len(values)))
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
		return fmt.Sprintf("%.6g", ff.Convert(reflect.TypeOf(float64(0))).Float())
	}
	// fallback by index
	if fallbackIndex >= 0 && fallbackIndex < f.NumField() {
		fld := f.Field(fallbackIndex)
		if fld.Kind() == reflect.Float32 || fld.Kind() == reflect.Float64 {
			return fmt.Sprintf("%.6g", fld.Convert(reflect.TypeOf(float64(0))).Float())
		}
	}
	return "<na>"
}
