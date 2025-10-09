package mygui

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"NR_Table/myast"
	"NR_Table/mydata"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	// テーブルのインデックスを表示するリスト（独立してスクロール可能）
	list := widget.NewList(
		func() int { return len(mydata.TableList) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("template")
			lbl.Wrapping = fyne.TextWrapOff
			return lbl
		},
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText("Table " + strconv.Itoa(i))
		},
	)

	// 新規: 外部の weldtbl.c を選択して読み込むボタン
	btnLoad := widget.NewButton("Load weldtbl...", func() {
		// モーダルでパス入力用のダイアログを表示
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Path to weldtbl.c")
		content := container.NewVBox(widget.NewLabel("Enter path to weldtbl.c:"), entry)
		dialog.ShowCustomConfirm("Load weldtbl", "Load", "Cancel", content, func(confirm bool) {
			if !confirm {
				return
			}
			path := entry.Text
			path = strings.Trim(path, `"`)
			if path == "" {
				dialog.ShowError(fmt.Errorf("invalid file path"), w)
				return
			}
			if err := myast.ParseWeldTable(path); err != nil {
				dialog.ShowError(err, w)
				return
			}
			// refresh list and select first table if any
			list.Refresh()
			if len(mydata.TableList) > 0 {
				list.Select(0)
			}
			dialog.ShowInformation("Loaded", "Loaded "+strconv.Itoa(len(mydata.TableList))+" tables", w)
		}, w)
	})
	topBtns := container.NewHBox(btnTwo, btnData, btnLoad)

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

	// 1) WeldCode: 折りたたみ式で表示（表示名は fieldDescriptions から取得）
	{
		title := fieldDescriptions["WeldCode"]
		grid := container.NewGridWithColumns(2)
		grid.Add(widget.NewLabel("Material"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.Material))))
		grid.Add(widget.NewLabel("Method"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.Method))))
		grid.Add(widget.NewLabel("PulseMode"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.PulseMode))))
		grid.Add(widget.NewLabel("PulseType"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.PulseType))))
		grid.Add(widget.NewLabel("Wire"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.Wire))))
		grid.Add(widget.NewLabel("Extension"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.Extension))))
		grid.Add(widget.NewLabel("Tip"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.Tip))))
		grid.Add(widget.NewLabel("Flag2"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.Flag2))))
		grid.Add(widget.NewLabel("Version"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.Version))))
		grid.Add(widget.NewLabel("StandardFlag"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.StandardFlag))))
		grid.Add(widget.NewLabel("Flag3"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.Flag3))))
		grid.Add(widget.NewLabel("LowSputter"))
		grid.Add(widget.NewLabel(strconv.Itoa(int(table.WeldCode.LowSputter))))

		// 折りたたみ式の Accordion として追加する
		item := widget.NewAccordionItem(title, grid)
		acc := widget.NewAccordion(item)
		sections = append(sections, acc, widget.NewSeparator())
	}

	// 2) WeldParm: 折りたたみ式で表示（504要素）
	{
		title := "WeldParm — " + fieldDescriptions["WeldParm"]
		item := widget.NewAccordionItem(title, createLazyGrid(func() fyne.CanvasObject {
			grid := container.NewGridWithColumns(2)
			for i := 0; i < len(table.WeldParm.Parm); i++ {
				num := strconv.Itoa(i + 1)
				for len(num) < 3 {
					num = "0" + num
				}
				grid.Add(widget.NewLabel("H" + num))
				hs := strings.ToUpper(strconv.FormatUint(uint64(table.WeldParm.Parm[i]), 16))
				for len(hs) < 4 {
					hs = "0" + hs
				}
				grid.Add(widget.NewLabel("0x" + hs))
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
					grid.Add(widget.NewLabel("V" + strconv.Itoa(i+1)))
					grid.Add(widget.NewLabel(a))
					grid.Add(widget.NewLabel(b))
					grid.Add(widget.NewLabel(c))
					grid.Add(widget.NewLabel(min))
					grid.Add(widget.NewLabel(max))
				}
				return grid
			} else if elemKind == reflect.Float32 || elemKind == reflect.Float64 {
				// フラット配列: 5個ずつグループ化して表示
				grid := container.NewGridWithColumns(6)
				grid.Add(widget.NewLabel("V#"))
				grid.Add(widget.NewLabel("a"))
				grid.Add(widget.NewLabel("b"))
				grid.Add(widget.NewLabel("c"))
				grid.Add(widget.NewLabel("min"))
				grid.Add(widget.NewLabel("max"))
				total := calcObj.Len()
				if total >= 5 {
					// グループ数は切り上げではなく、利用可能な要素で最大限表示する
					count := total / 5
					// if there is a remainder, still show last partial group
					if total%5 != 0 {
						count = (total + 4) / 5
					}
					for i := 0; i < count; i++ {
						base := i * 5
						grid.Add(widget.NewLabel("V" + strconv.Itoa(i+1)))
						for j := 0; j < 5; j++ {
							idx := base + j
							if idx < total {
								grid.Add(widget.NewLabel(strconv.FormatFloat(calcObj.Index(idx).Float(), 'g', 6, 64)))
							} else {
								grid.Add(widget.NewLabel("<na>"))
							}
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
					grid.Add(widget.NewLabel(strconv.Itoa(i)))
					elem := v.Index(i)
					// 型ごとに適切にフォーマットする
					var sval string
					switch elem.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						sval = strconv.FormatInt(elem.Int(), 10)
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
						sval = strconv.FormatUint(elem.Uint(), 10)
					case reflect.Float32, reflect.Float64:
						sval = strconv.FormatFloat(elem.Float(), 'g', 6, 64)
					case reflect.String:
						sval = elem.String()
					default:
						// fallback
						sval = fmt.Sprintf("%v", elem.Interface())
					}
					grid.Add(widget.NewLabel(sval))
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
		addArrayAccordion(vf.name+" — "+vf.desc, vf.arr)
	}

	// 6) CalParmDataTable: 折りたたみ式
	for idx := 0; idx < len(table.CalParmDataTable); idx++ {
		tbl := table.CalParmDataTable[idx]
		addArrayAccordion("CalParmDataTable["+strconv.Itoa(idx)+"] — "+fieldDescriptions["CalParmDataTable"], tbl.Data[:])
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
		return strconv.FormatFloat(f.Convert(reflect.TypeOf(float64(0))).Float(), 'g', 6, 64)
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
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Uint8:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', 6, 64)
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
