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
	"fyne.io/fyne/v2/widget"
)

// OpenDataCompare はデータ集計比較ウィンドウを開きます。
// 指示書に従い、指定したデータキーについて全テーブルの値を一覧し、
// 最も多い値をハイライトします（複数あればすべてハイライト）。
func OpenDataCompare(app fyne.App) {
	w := app.NewWindow("DataCompare")
	w.Resize(fyne.NewSize(900, 700))

	// 選択可能キーの用意（必要に応じて追加可）
	keys := []string{
		"WeldCode.Material",
		"WeldCode.Method",
		"WeldParm",  // requires index
		"A2S_Short", // requires index
		"S2V_Short", // requires index
		"CalParm",   // requires V index and subfield
		"V05_Data", "V06_Data", "V08_Data", "V12_Data",
		"V13_Data", "V15_Data", "V18_Data", "V19_Data", "V20_Data",
		"V32_Data", "V34_Data", "V36_Data",
		"V56_Data", "V59_Data", "V68_Data",
		"V94_Data", "V95_Data", "V57_Data", "V93_Data",
	}

	keySelect := widget.NewSelect(keys, nil)
	if len(keys) > 0 {
		keySelect.SetSelected(keys[0])
	}

	// index entry for array-like keys
	indexEntry := widget.NewEntry()
	indexEntry.SetPlaceHolder("index (0-based)")
	indexEntry.Disable()

	// calparm subfield selector (a,b,c,min,max) - enabled only if CalParm selected
	calSub := widget.NewSelect([]string{"a", "b", "c", "min", "max"}, nil)
	calSub.SetSelected("a")
	calSub.Disable()

	// status and result area
	status := widget.NewLabel("")
	resultBox := container.NewVBox()

	// helper to enable/disable controls based on selected key
	updateControls := func(selected string) {
		switch selected {
		case "WeldParm", "A2S_Short", "S2V_Short":
			indexEntry.Enable()
			calSub.Disable()
		case "CalParm":
			indexEntry.Enable() // expects V index
			calSub.Enable()
		default:
			indexEntry.Disable()
			calSub.Disable()
		}
	}

	keySelect.OnChanged = updateControls
	updateControls(keySelect.Selected)

	// value extractor: returns string representation for given table/key/index/sub
	getValue := func(tbl *mydata.TableData, key string, idx int, sub string) string {
		v := reflect.ValueOf(*tbl)
		switch key {
		case "WeldCode.Material":
			return fmt.Sprintf("%d", tbl.WeldCode.Material)
		case "WeldCode.Method":
			return fmt.Sprintf("%d", tbl.WeldCode.Method)
		case "WeldParm":
			if idx >= 0 && idx < len(tbl.WeldParm.Parm) {
				return fmt.Sprintf("0x%04X", tbl.WeldParm.Parm[idx])
			}
			return "<oob>"
		case "A2S_Short":
			if idx >= 0 && idx < len(tbl.A2S_Short.Speed) {
				return fmt.Sprintf("%d", tbl.A2S_Short.Speed[idx])
			}
			return "<oob>"
		case "S2V_Short":
			if idx >= 0 && idx < len(tbl.S2V_Short.Values) {
				return fmt.Sprintf("%d", tbl.S2V_Short.Values[idx])
			}
			return "<oob>"
		case "CalParm":
			// CalParm may be either [116]float32 (flat) or []DCCALPARM-like struct.
			rv := v.FieldByName("CalParm")
			if !rv.IsValid() {
				return "<na>"
			}
			// if array/slice of structs
			if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
				elem := rv.Index(0)
				if elem.Kind() == reflect.Struct {
					if idx >= 0 && idx < rv.Len() {
						el := rv.Index(idx)
						switch sub {
						case "a", "A":
							return formatReflectFloatField(el, "A", 0)
						case "b", "B":
							return formatReflectFloatField(el, "B", 1)
						case "c", "C":
							return formatReflectFloatField(el, "C", 2)
						case "min", "Min":
							return formatReflectFloatField(el, "Min", 3)
						case "max", "Max":
							return formatReflectFloatField(el, "Max", 4)
						}
					}
					return "<oob>"
				}
				// elements are floats (flat representation), group by 5-per-V
				if elem.Kind() == reflect.Float32 || elem.Kind() == reflect.Float64 {
					// treat CalParm as flattened floats; idx is Vn-1, sub selects offset within 5
					base := idx * 5
					offset := 0
					switch sub {
					case "a":
						offset = 0
					case "b":
						offset = 1
					case "c":
						offset = 2
					case "min":
						offset = 3
					case "max":
						offset = 4
					}
					if base+offset >= 0 && base+offset < rv.Len() {
						fv := rv.Index(base + offset).Float()
						return fmt.Sprintf("%.6g", fv)
					}
					return "<oob>"
				}
			}
			return "<unknown CalParm>"
		default:
			// try to map Vxx fields by reflection
			f := v.FieldByName(key)
			if f.IsValid() && (f.Kind() == reflect.Array || f.Kind() == reflect.Slice) {
				if idx >= 0 && idx < f.Len() {
					return formatValue(f.Index(idx))
				}
				return "<oob>"
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
		key := keySelect.Selected
		if key == "" {
			status.SetText("Select a key")
			return
		}
		idx := -1
		if indexEntry.Text != "" {
			if n, err := strconv.Atoi(indexEntry.Text); err == nil {
				idx = n
			} else {
				status.SetText("Invalid index")
				return
			}
		}
		sub := calSub.Selected

		counts := map[string]int{}
		values := make([]string, len(mydata.TableList))
		for i := range mydata.TableList {
			val := getValue(&mydata.TableList[i], key, idx, sub)
			values[i] = val
			counts[val]++
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
		// replace result content with a taller, scrollable table (no summary/legend below)
		resultBox.Objects = nil

		// create a grid with 2 columns: Table index, Value
		grid := container.NewGridWithColumns(2)
		// header
		grid.Add(widget.NewLabelWithStyle("Table", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		grid.Add(widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		for i, val := range values {
			grid.Add(widget.NewLabel(fmt.Sprintf("%d", i)))
			col := color.RGBA{R: 0, G: 0, B: 0, A: 255}
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

	controls := container.NewHBox(
		widget.NewLabel("Key:"), keySelect,
		widget.NewLabel("Index:"), indexEntry,
		widget.NewLabel("Cal sub:"), calSub,
		compareBtn, status,
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
