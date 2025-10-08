package mygui

import (
	"fmt"
	"reflect"

	"NR_Table/mydata"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// メインウィンドウ
type Mw struct {
	window fyne.Window
	uiItem MwUiItem
}

// メインウィンドウに配置するUIパーツ
type MwUiItem struct {
	tableList *widget.List
	detail    fyne.CanvasObject
}

// NewMw はメインウィンドウとUIを作成します
func NewMw(app fyne.App) *Mw {
	w := app.NewWindow("NR Table v0.1")
	w.Resize(fyne.NewSize(1000, 700))

	// 左側：上部にボタン、残りの縦領域を占めるリスト
	btnTwo := widget.NewButton("Open TwoCompare", func() {
		OpenTwoCompare(app)
	})
	btnData := widget.NewButton("Open DataCompare", func() {
		OpenDataCompare(app)
	})
	topBtns := container.NewHBox(btnTwo, btnData)

	// テーブルのインデックスを表示するリスト（独立してスクロール可能）
	list := widget.NewList(
		func() int { return len(mydata.TableList) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("template")
			lbl.Wrapping = fyne.TextWrapOff
			return lbl
		},
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fmt.Sprintf("Table %d", i))
		},
	)

	// 右側：初期表示（テーブル選択を促すヒント）
	rightContent := container.NewVBox(widget.NewLabel("Select a table from the list."))
	rightScroll := container.NewVScroll(rightContent)
	rightScroll.SetMinSize(fyne.NewSize(600, 400))

	// 左ペインを構築：上部にボタン、残りをリストが埋める
	leftPane := container.NewBorder(topBtns, nil, nil, nil, list)

	// メインの分割：左（ボタン＋リスト）、右（詳細）。左は初期幅30%。
	split := container.NewHSplit(leftPane, rightScroll)
	split.SetOffset(0.3)

	// リスト選択処理：右ペインにテーブル全体表示を設定（縦スクロールのみ）
	list.OnSelected = func(i int) {
		if i < 0 || i >= len(mydata.TableList) {
			rightContent.Objects = []fyne.CanvasObject{widget.NewLabel("Out of range")}
			rightContent.Refresh()
			return
		}
		t := mydata.TableList[i] // copy
		content := buildDetailView(&t)
		// 右ペインのコンテナ内容を置き換える（外側の rightScroll が縦スクロールを担う）
		rightContent.Objects = []fyne.CanvasObject{content}
		rightContent.Refresh()
	}

	w.SetContent(split)

	return &Mw{
		window: w,
		uiItem: MwUiItem{
			tableList: list,
			detail:    rightScroll,
		},
	}
}

// ShowAndRun はメインウィンドウを表示してアプリを実行します
func (mw *Mw) ShowAndRun() {
	mw.window.ShowAndRun()
}

// フィールド名 -> 表示用説明（指示書に基づく）
var fieldDescriptions = map[string]string{
	"WeldCode":      "溶接種別コード",
	"A2S_Pulse":     "ワイヤ送給テーブル（パルス）",
	"S2V_Pulse":     "一元電圧テーブル（パルス）",
	"A2S_Short":     "ワイヤ送給テーブル（短絡）",
	"S2V_Short":     "一元電圧テーブル（短絡）",
	"WeldParm":      "半固定パラメータ",
	"CalParm":       "可変パラメータ係数",
	"ParmTbl_Pls":   "パラメータテーブル（パルス）",
	"ParmTbl_Short": "パラメータテーブル（短絡）",
	"CalParmList":   "可変パラメータのテーブル引きリスト",
	// Vxx names
	"V05_Data":         "V5 テーブル",
	"V06_Data":         "V6 テーブル",
	"V08_Data":         "V8 テーブル",
	"V12_Data":         "V12 テーブル",
	"V32_Data":         "V32 テーブル",
	"V34_Data":         "V34 テーブル",
	"V36_Data":         "V36 テーブル",
	"V56_Data":         "V56 テーブル",
	"V59_Data":         "V59 テーブル",
	"V68_Data":         "V68 テーブル",
	"V13_Data":         "V13 テーブル",
	"V15_Data":         "V15 テーブル",
	"V18_Data":         "V18 テーブル",
	"V19_Data":         "V19 テーブル",
	"V20_Data":         "V20 テーブル",
	"V94_Data":         "V94 テーブル",
	"V95_Data":         "V95 テーブル",
	"V57_Data":         "V57 テーブル",
	"V93_Data":         "V93 テーブル",
	"CalParmDataTable": "可変パラメータのテーブル引きデータ",
	"Navi_Pram1":       "短絡用溶接ナビデータ：T継ぎ手データ",
	"Navi_Pram2":       "短絡用溶接ナビデータ：重ね継ぎ手データ",
	"Navi_Pram3":       "短絡用溶接ナビデータ：突き合わせデータ",
	"Navi_P_Pram1":     "パルス用溶接ナビデータ：T継ぎ手データ",
	"Navi_P_Pram2":     "パルス用溶接ナビデータ：重ね継ぎ手データ",
	"Navi_P_Pram3":     "パルス用溶接ナビデータ：突き合わせデータ",
}

// buildDetailView は TableData を縦方向のセクション（ヘッダ＋表）で表示するコンテンツを返します。
// 横スクロールは禁止、すべて縦方向に展開します。
// 高速化のため、各セクションを折りたたみ式（Accordion）で表示します。
func buildDetailView(table *mydata.TableData) fyne.CanvasObject {
	sections := []fyne.CanvasObject{}

	// 1) WeldCode: key/value 表（軽量なので常時表示）
	{
		secLabel := widget.NewLabel("WeldCode — " + fieldDescriptions["WeldCode"])
		secLabel.TextStyle = fyne.TextStyle{Bold: true}
		grid := container.NewGridWithColumns(2)
		grid.Add(widget.NewLabel("Material"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.Material)))
		grid.Add(widget.NewLabel("Method"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.Method)))
		grid.Add(widget.NewLabel("PulseMode"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.PulseMode)))
		grid.Add(widget.NewLabel("PulseType"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.PulseType)))
		grid.Add(widget.NewLabel("Wire"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.Wire)))
		grid.Add(widget.NewLabel("Extension"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.Extension)))
		grid.Add(widget.NewLabel("Tip"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.Tip)))
		grid.Add(widget.NewLabel("Flag2"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.Flag2)))
		grid.Add(widget.NewLabel("Version"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.Version)))
		grid.Add(widget.NewLabel("StandardFlag"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.StandardFlag)))
		grid.Add(widget.NewLabel("Flag3"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.Flag3)))
		grid.Add(widget.NewLabel("LowSputter"))
		grid.Add(widget.NewLabel(fmt.Sprintf("%d", table.WeldCode.LowSputter)))
		sections = append(sections, secLabel, grid, widget.NewSeparator())
	}

	// 2) WeldParm: 折りたたみ式で表示（504要素）
	{
		title := "WeldParm — " + fieldDescriptions["WeldParm"]
		item := widget.NewAccordionItem(title, createLazyGrid(func() fyne.CanvasObject {
			grid := container.NewGridWithColumns(2)
			for i := 0; i < len(table.WeldParm.Parm); i++ {
				grid.Add(widget.NewLabel(fmt.Sprintf("H%03d", i+1)))
				grid.Add(widget.NewLabel(fmt.Sprintf("0x%04X", table.WeldParm.Parm[i])))
			}
			return grid
		}))
		acc := widget.NewAccordion(item)
		sections = append(sections, acc, widget.NewSeparator())
	}

	// 3) CalParm: 折りたたみ式で表示
	{
		title := "CalParm — " + fieldDescriptions["CalParm"]
		item := widget.NewAccordionItem(title, createLazyGrid(func() fyne.CanvasObject {
			calcObj := reflect.ValueOf(table.CalParm)
			elemKind := calcObj.Type().Elem().Kind()

			if elemKind == reflect.Struct {
				grid := container.NewGridWithColumns(6)
				grid.Add(widget.NewLabel("V#"))
				grid.Add(widget.NewLabel("a"))
				grid.Add(widget.NewLabel("b"))
				grid.Add(widget.NewLabel("c"))
				grid.Add(widget.NewLabel("min"))
				grid.Add(widget.NewLabel("max"))
				for i := 0; i < calcObj.Len(); i++ {
					el := calcObj.Index(i)
					a := getFloatFieldAsString(el, "A")
					if a == "" {
						a = formatValue(el.Field(0))
					}
					b := getFloatFieldAsString(el, "B")
					if b == "" {
						b = formatValue(el.Field(1))
					}
					c := getFloatFieldAsString(el, "C")
					if c == "" {
						c = formatValue(el.Field(2))
					}
					min := getFloatFieldAsString(el, "Min")
					if min == "" {
						min = formatValue(el.Field(3))
					}
					max := getFloatFieldAsString(el, "Max")
					if max == "" {
						max = formatValue(el.Field(4))
					}
					grid.Add(widget.NewLabel(fmt.Sprintf("V%d", i+1)))
					grid.Add(widget.NewLabel(a))
					grid.Add(widget.NewLabel(b))
					grid.Add(widget.NewLabel(c))
					grid.Add(widget.NewLabel(min))
					grid.Add(widget.NewLabel(max))
				}
				return grid
			} else if elemKind == reflect.Float32 || elemKind == reflect.Float64 {
				grid := container.NewGridWithColumns(6)
				grid.Add(widget.NewLabel("V#"))
				grid.Add(widget.NewLabel("a"))
				grid.Add(widget.NewLabel("b"))
				grid.Add(widget.NewLabel("c"))
				grid.Add(widget.NewLabel("min"))
				grid.Add(widget.NewLabel("max"))
				total := calcObj.Len()
				if total >= 5 && total%5 == 0 {
					count := total / 5
					for i := 0; i < count; i++ {
						base := i * 5
						grid.Add(widget.NewLabel(fmt.Sprintf("V%d", i+1)))
						for j := 0; j < 5; j++ {
							grid.Add(widget.NewLabel(fmt.Sprintf("%.6g", calcObj.Index(base+j).Float())))
						}
					}
				}
				return grid
			}
			return widget.NewLabel("Unknown CalParm format")
		}))
		acc := widget.NewAccordion(item)
		sections = append(sections, acc, widget.NewSeparator())
	}

	// 4) A2S / S2V テーブル: 折りたたみ式（各256要素）
	addArrayAccordion := func(title string, data interface{}) {
		item := widget.NewAccordionItem(title, createLazyGrid(func() fyne.CanvasObject {
			grid := container.NewGridWithColumns(2)
			v := reflect.ValueOf(data)
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				for i := 0; i < v.Len(); i++ {
					grid.Add(widget.NewLabel(fmt.Sprintf("%d", i)))
					grid.Add(widget.NewLabel(fmt.Sprintf("%v", v.Index(i).Interface())))
				}
			}
			return grid
		}))
		acc := widget.NewAccordion(item)
		sections = append(sections, acc, widget.NewSeparator())
	}

	addArrayAccordion("A2S_Pulse — "+fieldDescriptions["A2S_Pulse"], table.A2S_Pulse.Speed[:])
	addArrayAccordion("S2V_Pulse — "+fieldDescriptions["S2V_Pulse"], table.S2V_Pulse.Values[:])
	addArrayAccordion("A2S_Short — "+fieldDescriptions["A2S_Short"], table.A2S_Short.Speed[:])
	addArrayAccordion("S2V_Short — "+fieldDescriptions["S2V_Short"], table.S2V_Short.Values[:])

	// 5) Vxx データ: 折りたたみ式（各128要素）
	vfields := []struct {
		name string
		arr  []int16
		desc string
	}{
		{"V05_Data", table.V05_Data[:], fieldDescriptions["V05_Data"]},
		{"V06_Data", table.V06_Data[:], fieldDescriptions["V06_Data"]},
		{"V08_Data", table.V08_Data[:], fieldDescriptions["V08_Data"]},
		{"V12_Data", table.V12_Data[:], fieldDescriptions["V12_Data"]},
		{"V32_Data", table.V32_Data[:], fieldDescriptions["V32_Data"]},
		{"V34_Data", table.V34_Data[:], fieldDescriptions["V34_Data"]},
		{"V36_Data", table.V36_Data[:], fieldDescriptions["V36_Data"]},
		{"V56_Data", table.V56_Data[:], fieldDescriptions["V56_Data"]},
		{"V59_Data", table.V59_Data[:], fieldDescriptions["V59_Data"]},
		{"V68_Data", table.V68_Data[:], fieldDescriptions["V68_Data"]},
		{"V13_Data", table.V13_Data[:], fieldDescriptions["V13_Data"]},
		{"V15_Data", table.V15_Data[:], fieldDescriptions["V15_Data"]},
		{"V18_Data", table.V18_Data[:], fieldDescriptions["V18_Data"]},
		{"V19_Data", table.V19_Data[:], fieldDescriptions["V19_Data"]},
		{"V20_Data", table.V20_Data[:], fieldDescriptions["V20_Data"]},
		{"V94_Data", table.V94_Data[:], fieldDescriptions["V94_Data"]},
		{"V95_Data", table.V95_Data[:], fieldDescriptions["V95_Data"]},
		{"V57_Data", table.V57_Data[:], fieldDescriptions["V57_Data"]},
		{"V93_Data", table.V93_Data[:], fieldDescriptions["V93_Data"]},
	}
	for _, vf := range vfields {
		addArrayAccordion(fmt.Sprintf("%s — %s", vf.name, vf.desc), vf.arr)
	}

	// 6) CalParmDataTable: 折りたたみ式
	for idx := 0; idx < len(table.CalParmDataTable); idx++ {
		tbl := table.CalParmDataTable[idx]
		addArrayAccordion(fmt.Sprintf("CalParmDataTable[%d] — %s", idx, fieldDescriptions["CalParmDataTable"]), tbl.Data[:])
	}

	// 7) Navi arrays: 折りたたみ式（各7要素、軽量なので展開しても可）
	addArrayAccordion("Navi_Pram1 — "+fieldDescriptions["Navi_Pram1"], table.Navi_Pram1[:])
	addArrayAccordion("Navi_Pram2 — "+fieldDescriptions["Navi_Pram2"], table.Navi_Pram2[:])
	addArrayAccordion("Navi_Pram3 — "+fieldDescriptions["Navi_Pram3"], table.Navi_Pram3[:])
	addArrayAccordion("Navi_P_Pram1 — "+fieldDescriptions["Navi_P_Pram1"], table.Navi_P_Pram1[:])
	addArrayAccordion("Navi_P_Pram2 — "+fieldDescriptions["Navi_P_Pram2"], table.Navi_P_Pram2[:])
	addArrayAccordion("Navi_P_Pram3 — "+fieldDescriptions["Navi_P_Pram3"], table.Navi_P_Pram3[:])

	// コンテンツを縦方向にまとめる
	content := container.NewVBox(sections...)
	return content
}

// createLazyGrid は遅延レンダリング用のコンテナを返します
func createLazyGrid(buildFunc func() fyne.CanvasObject) fyne.CanvasObject {
	// プレースホルダを返し、展開時に実際のコンテンツを生成します
	placeholder := widget.NewLabel("クリックして展開...")
	var actualContent fyne.CanvasObject
	rendered := false

	box := container.NewVBox()
	btn := widget.NewButton("表示", func() {
		if rendered {
			return
		}
		actualContent = buildFunc()
		// replace placeholder with actual content
		if len(box.Objects) >= 2 {
			box.Objects[1] = actualContent
		} else {
			box.Add(actualContent)
		}
		box.Refresh()
		rendered = true
	})
	box.Add(btn)
	box.Add(placeholder)
	return box
}

// getFloatFieldAsString は reflect.Struct のフィールド名で float を取り出す簡易ヘルパ
func getFloatFieldAsString(v reflect.Value, fieldName string) string {
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ""
	}
	f := v.FieldByName(fieldName)
	if f.IsValid() && (f.Kind() == reflect.Float32 || f.Kind() == reflect.Float64) {
		return fmt.Sprintf("%.6g", f.Convert(reflect.TypeOf(float64(0))).Float())
	}
	return ""
}

// formatValue は reflect.Value を可読文字列に変換します
func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Uint8:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.6g", v.Float())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
