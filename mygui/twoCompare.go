package mygui

import (
	"fmt"
	"image/color"
	"reflect"
	"strconv"
	"strings"

	"NR_Table/mydata"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// makeWrappedLabel は折り返し有効なラベルを生成し、表示が縦並びにならないよう最小幅を指定します。
// makeWrappedLabel は折り返し有効なラベルを生成します。
// ラベル単体では最小幅を設定できないため、必要に応じて
// wrapWithMinWidth でラベルを包んで使用してください。
func makeWrappedLabel(text string) *widget.Label {
	lbl := widget.NewLabel(text)
	lbl.Wrapping = fyne.TextWrapWord
	return lbl
}

// wrapWithMinWidth は与えたラベルを横幅を確保するコンテナで包みます。
// これにより長い単語が縦に1文字ずつ並ぶ縦書きのような表示を防ぎます。
func wrapWithMinWidth(c fyne.CanvasObject, minWidth float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(minWidth, 1))
	return container.NewHBox(spacer, c)
}

// OpenTwoCompare は 2 つのテーブルを比較するウィンドウを開きます。
// 左右にテーブルを並べ、差分を色でハイライトします。
// 全データを表形式で表示し、大量データは遅延レンダリングで高速化します。
func OpenTwoCompare(app fyne.App) {
	w := app.NewWindow("TwoCompare")
	w.Resize(fyne.NewSize(1200, 700))

	// テーブル数チェック
	numTables := len(mydata.TableList)
	if numTables == 0 {
		w.SetContent(container.NewVBox(widget.NewLabel("No tables loaded")))
		w.Show()
		return
	}

	// インデックスリスト作成
	idxs := make([]string, numTables)
	for i := 0; i < numTables; i++ {
		idxs[i] = strconv.Itoa(i)
	}

	// UI パーツ
	leftSelect := widget.NewSelect(idxs, nil)
	rightSelect := widget.NewSelect(idxs, nil)
	leftSelect.SetSelectedIndex(0)
	if len(idxs) > 1 {
		rightSelect.SetSelectedIndex(1)
	} else {
		rightSelect.SetSelectedIndex(0)
	}

	compareBtn := widget.NewButton("Compare", nil)
	status := widget.NewLabel("")

	// 表示領域（縦スクロールのみ）
	// 初期は左側用の空のコンテンツを作る（後でウィンドウ中央を差し替える）
	leftContent := container.NewVBox(widget.NewLabel("Left table not selected"))
	leftScroll := container.NewVScroll(leftContent)

	// コントロール行（先に作成して compareAction で使用する）
	controls := container.NewHBox(
		widget.NewLabel("Left:"), leftSelect,
		widget.NewLabel("Right:"), rightSelect,
		compareBtn, status,
	)

	// 比較処理
	compareAction := func() {
		ls := leftSelect.Selected
		rs := rightSelect.Selected
		if ls == "" || rs == "" {
			status.SetText("Select both tables")
			return
		}
		li, err1 := strconv.Atoi(ls)
		ri, err2 := strconv.Atoi(rs)
		if err1 != nil || err2 != nil || li < 0 || ri < 0 || li >= len(mydata.TableList) || ri >= len(mydata.TableList) {
			status.SetText("Invalid indices")
			return
		}
		a := mydata.TableList[li]
		b := mydata.TableList[ri]

		// 左右を統合した比較ビューを作成（同期スクロール）
		unifiedView := buildUnifiedComparisonView(&a, &b)

		// 新しいスクロールを作成（1つのスクロールで左右を包む）
		newScroll := container.NewVScroll(unifiedView)

		// ウィンドウの中央コンテンツを新しいスクロールに置き換える
		newContent := container.NewBorder(controls, nil, nil, nil, newScroll)
		w.SetContent(newContent)

		newScroll.ScrollToTop()

		status.SetText("Compared Table " + strconv.Itoa(li) + " ↔ Table " + strconv.Itoa(ri))
	}

	compareBtn.OnTapped = compareAction

	// メインレイアウト: 上部にコントロール、下部は中央に単一スクロール
	content := container.NewBorder(controls, nil, nil, nil, leftScroll)
	w.SetContent(content)
	w.Show()
}

// buildUnifiedComparisonView は左右のテーブルを並べて表示する統合ビューを構築します。
// 1つのスクロール領域に左右を配置することで同期スクロールを実現します。
func buildUnifiedComparisonView(a, b *mydata.TableData) fyne.CanvasObject {
	sections := []fyne.CanvasObject{}

	va := reflect.ValueOf(*a)
	vb := reflect.ValueOf(*b)
	ta := va.Type()

	for i := 0; i < ta.NumField(); i++ {
		f := ta.Field(i)
		name := f.Name
		desc := fieldDescriptions[name]
		header := name
		if desc != "" {
			header = name + " — " + desc
		}

		// セクション見出し（折り返し有効）
		h := widget.NewLabelWithStyle(header, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		h.Wrapping = fyne.TextWrapWord
		sections = append(sections, h)

		fa := va.Field(i)
		fb := vb.Field(i)

		// WeldParm は専用表 (H###, 0xXXXX) - 左右3列で表示
		if name == "WeldParm" && fa.Kind() == reflect.Struct {
			item := widget.NewAccordionItem(header, createLazyGrid(func() fyne.CanvasObject {
				arr := fa.FieldByName("Parm")
				arrB := fb.FieldByName("Parm")
				if !arr.IsValid() || (arr.Kind() != reflect.Array && arr.Kind() != reflect.Slice) {
					return widget.NewLabel("WeldParm format unsupported")
				}
				n := arr.Len()
				grid := container.NewGridWithColumns(3)
				grid.Add(widget.NewLabel("H#"))
				grid.Add(widget.NewLabel("Left"))
				grid.Add(widget.NewLabel("Right"))
				for j := 0; j < n; j++ {
					// H### (zero-padded)
					num := strconv.Itoa(j + 1)
					for len(num) < 3 {
						num = "0" + num
					}
					grid.Add(widget.NewLabel("H" + num))
					// 0x + 4桁大文字16進
					hv := arr.Index(j).Uint()
					hs := strings.ToUpper(strconv.FormatUint(hv, 16))
					for len(hs) < 4 {
						hs = "0" + hs
					}
					leftStr := "0x" + hs
					rightStr := "<nil>"
					if arrB.IsValid() && j < arrB.Len() {
						hvb := arrB.Index(j).Uint()
						hsb := strings.ToUpper(strconv.FormatUint(hvb, 16))
						for len(hsb) < 4 {
							hsb = "0" + hsb
						}
						rightStr = "0x" + hsb
					}

					lbl1 := wrapWithMinWidth(makeWrappedLabel(leftStr), 260)
					lbl2 := wrapWithMinWidth(makeWrappedLabel(rightStr), 260)

					c := pickCompareColor(leftStr, rightStr)
					if c != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
						rect1 := canvas.NewRectangle(c)
						rect1.SetMinSize(fyne.NewSize(10, 10))
						rect2 := canvas.NewRectangle(c)
						rect2.SetMinSize(fyne.NewSize(10, 10))
						grid.Add(container.NewHBox(rect1, lbl1))
						grid.Add(container.NewHBox(rect2, lbl2))
					} else {
						grid.Add(lbl1)
						grid.Add(lbl2)
					}
				}
				return grid
			}))
			acc := widget.NewAccordion(item)
			sections = append(sections, acc, widget.NewSeparator())
			continue
		}

		// CalParm は専用表 (V#, a,b,c,min,max) - 左右で並べる
		if name == "CalParm" {
			item := widget.NewAccordionItem(header, createLazyGrid(func() fyne.CanvasObject {
				calcObjA := fa
				calcObjB := fb
				if calcObjA.Kind() != reflect.Array && calcObjA.Kind() != reflect.Slice {
					return widget.NewLabel("CalParm format unsupported")
				}
				elemKind := calcObjA.Type().Elem().Kind()
				if elemKind == reflect.Struct {
					// 構造体配列: V# | Left(a,b,c,min,max) | Right(a,b,c,min,max)
					vbox := container.NewVBox()
					headerRow := container.NewGridWithColumns(3)
					headerRow.Add(widget.NewLabel("V#"))
					headerRow.Add(widget.NewLabel("Left (a,b,c,min,max)"))
					headerRow.Add(widget.NewLabel("Right (a,b,c,min,max)"))
					vbox.Add(headerRow)

					for vi := 0; vi < calcObjA.Len(); vi++ {
						elA := calcObjA.Index(vi)
						elB := reflect.Value{}
						if vi < calcObjB.Len() {
							elB = calcObjB.Index(vi)
						}

						// 左側の値
						aStr := getFloatFieldAsString(elA, "A")
						if aStr == "" {
							aStr = formatCompareValue(elA.Field(0))
						}
						bStr := getFloatFieldAsString(elA, "B")
						if bStr == "" {
							bStr = formatCompareValue(elA.Field(1))
						}
						cStr := getFloatFieldAsString(elA, "C")
						if cStr == "" {
							cStr = formatCompareValue(elA.Field(2))
						}
						minStr := getFloatFieldAsString(elA, "Min")
						if minStr == "" {
							minStr = formatCompareValue(elA.Field(3))
						}
						maxStr := getFloatFieldAsString(elA, "Max")
						if maxStr == "" {
							maxStr = formatCompareValue(elA.Field(4))
						}
						leftVal := strings.Join([]string{aStr, bStr, cStr, minStr, maxStr}, ",")

						// 右側の値
						var aStrB, bStrB, cStrB, minStrB, maxStrB string
						if elB.IsValid() {
							aStrB = getFloatFieldAsString(elB, "A")
							if aStrB == "" {
								aStrB = formatCompareValue(elB.Field(0))
							}
							bStrB = getFloatFieldAsString(elB, "B")
							if bStrB == "" {
								bStrB = formatCompareValue(elB.Field(1))
							}
							cStrB = getFloatFieldAsString(elB, "C")
							if cStrB == "" {
								cStrB = formatCompareValue(elB.Field(2))
							}
							minStrB = getFloatFieldAsString(elB, "Min")
							if minStrB == "" {
								minStrB = formatCompareValue(elB.Field(3))
							}
							maxStrB = getFloatFieldAsString(elB, "Max")
							if maxStrB == "" {
								maxStrB = formatCompareValue(elB.Field(4))
							}
						}
						rightVal := strings.Join([]string{aStrB, bStrB, cStrB, minStrB, maxStrB}, ",")

						row := container.NewGridWithColumns(3)
						row.Add(widget.NewLabel("V" + strconv.Itoa(vi+1)))

						lbl1 := wrapWithMinWidth(makeWrappedLabel(leftVal), 260)
						lbl2 := wrapWithMinWidth(makeWrappedLabel(rightVal), 260)

						c := pickCompareColor(leftVal, rightVal)
						if c != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
							rect1 := canvas.NewRectangle(c)
							rect1.SetMinSize(fyne.NewSize(10, 10))
							rect2 := canvas.NewRectangle(c)
							rect2.SetMinSize(fyne.NewSize(10, 10))
							row.Add(container.NewHBox(rect1, lbl1))
							row.Add(container.NewHBox(rect2, lbl2))
						} else {
							row.Add(lbl1)
							row.Add(lbl2)
						}
						vbox.Add(row)
					}
					return vbox
				} else if elemKind == reflect.Float32 || elemKind == reflect.Float64 {
					// フラット配列: 5個ずつグループ化
					vbox := container.NewVBox()
					headerRow := container.NewGridWithColumns(3)
					headerRow.Add(widget.NewLabel("V#"))
					headerRow.Add(widget.NewLabel("Left (a,b,c,min,max)"))
					headerRow.Add(widget.NewLabel("Right (a,b,c,min,max)"))
					vbox.Add(headerRow)

					total := calcObjA.Len()
					if total >= 5 && total%5 == 0 {
						count := total / 5
						for vi := 0; vi < count; vi++ {
							base := vi * 5
							leftVals := make([]string, 5)
							rightVals := make([]string, 5)
							for j := 0; j < 5; j++ {
								leftVals[j] = strconv.FormatFloat(calcObjA.Index(base+j).Float(), 'g', 6, 64)
								if base+j < calcObjB.Len() {
									rightVals[j] = strconv.FormatFloat(calcObjB.Index(base+j).Float(), 'g', 6, 64)
								} else {
									rightVals[j] = "<nil>"
								}
							}
							leftStr := strings.Join(leftVals, ",")
							rightStr := strings.Join(rightVals, ",")

							row := container.NewGridWithColumns(3)
							row.Add(widget.NewLabel("V" + strconv.Itoa(vi+1)))

							lbl1 := wrapWithMinWidth(makeWrappedLabel(leftStr), 260)
							lbl2 := wrapWithMinWidth(makeWrappedLabel(rightStr), 260)

							c := pickCompareColor(leftStr, rightStr)
							if c != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
								rect1 := canvas.NewRectangle(c)
								rect1.SetMinSize(fyne.NewSize(10, 10))
								rect2 := canvas.NewRectangle(c)
								rect2.SetMinSize(fyne.NewSize(10, 10))
								row.Add(container.NewHBox(rect1, lbl1))
								row.Add(container.NewHBox(rect2, lbl2))
							} else {
								row.Add(lbl1)
								row.Add(lbl2)
							}
							vbox.Add(row)
						}
					}
					return vbox
				}
				return widget.NewLabel("Unknown CalParm format")
			}))
			acc := widget.NewAccordion(item)
			sections = append(sections, acc, widget.NewSeparator())
			continue
		}

		// 配列／スライス (汎用): Index | Left | Right の3列
		if fa.Kind() == reflect.Array || fa.Kind() == reflect.Slice {
			title := header + " (array)"
			item := widget.NewAccordionItem(title, createLazyGrid(func() fyne.CanvasObject {
				grid := container.NewGridWithColumns(3)
				grid.Add(widget.NewLabel("Index"))
				grid.Add(widget.NewLabel("Left"))
				grid.Add(widget.NewLabel("Right"))
				n := fa.Len()
				for idx := 0; idx < n; idx++ {
					grid.Add(widget.NewLabel(strconv.Itoa(idx)))
					leftStr := formatCompareValue(fa.Index(idx))
					rightStr := "<nil>"
					if idx < fb.Len() {
						rightStr = formatCompareValue(fb.Index(idx))
					}

					lbl1 := wrapWithMinWidth(makeWrappedLabel(leftStr), 260)
					lbl2 := wrapWithMinWidth(makeWrappedLabel(rightStr), 260)

					c := pickCompareColor(leftStr, rightStr)
					if c != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
						rect1 := canvas.NewRectangle(c)
						rect1.SetMinSize(fyne.NewSize(10, 10))
						rect2 := canvas.NewRectangle(c)
						rect2.SetMinSize(fyne.NewSize(10, 10))
						grid.Add(container.NewHBox(rect1, lbl1))
						grid.Add(container.NewHBox(rect2, lbl2))
					} else {
						grid.Add(lbl1)
						grid.Add(lbl2)
					}
				}
				return grid
			}))
			acc := widget.NewAccordion(item)
			sections = append(sections, acc, widget.NewSeparator())
			continue
		}

		// 構造体: Field | Left | Right の3列
		if fa.Kind() == reflect.Struct {
			grid := container.NewGridWithColumns(3)
			grid.Add(widget.NewLabel("Field"))
			grid.Add(widget.NewLabel("Left"))
			grid.Add(widget.NewLabel("Right"))
			for j := 0; j < fa.NumField(); j++ {
				subName := fa.Type().Field(j).Name
				grid.Add(widget.NewLabel(subName))
				leftStr := formatCompareValue(fa.Field(j))
				rightStr := formatCompareValue(fb.Field(j))

				lbl1 := wrapWithMinWidth(makeWrappedLabel(leftStr), 260)
				lbl2 := wrapWithMinWidth(makeWrappedLabel(rightStr), 260)

				c := pickCompareColor(leftStr, rightStr)
				if c != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
					rect1 := canvas.NewRectangle(c)
					rect1.SetMinSize(fyne.NewSize(10, 10))
					rect2 := canvas.NewRectangle(c)
					rect2.SetMinSize(fyne.NewSize(10, 10))
					grid.Add(container.NewHBox(rect1, lbl1))
					grid.Add(container.NewHBox(rect2, lbl2))
				} else {
					grid.Add(lbl1)
					grid.Add(lbl2)
				}
			}
			sections = append(sections, grid, widget.NewSeparator())
			continue
		}

		// プリミティブ値: Left | Right の2列
		leftStr := formatCompareValue(fa)
		rightStr := formatCompareValue(fb)

		row := container.NewGridWithColumns(2)
		lbl1 := wrapWithMinWidth(makeWrappedLabel(leftStr), 260)
		lbl2 := wrapWithMinWidth(makeWrappedLabel(rightStr), 260)

		c := pickCompareColor(leftStr, rightStr)
		if c != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
			rect1 := canvas.NewRectangle(c)
			rect1.SetMinSize(fyne.NewSize(10, 10))
			rect2 := canvas.NewRectangle(c)
			rect2.SetMinSize(fyne.NewSize(10, 10))
			row.Add(container.NewHBox(rect1, lbl1))
			row.Add(container.NewHBox(rect2, lbl2))
		} else {
			row.Add(lbl1)
			row.Add(lbl2)
		}
		sections = append(sections, row, widget.NewSeparator())
	}

	return container.NewVBox(sections...)
}

// pickCompareColor は a と b が異なるとき赤、同じなら黒を返します。
func pickCompareColor(a, b string) color.Color {
	if a != b {
		return color.RGBA{R: 200, G: 0, B: 0, A: 255}
	}
	return color.RGBA{R: 0, G: 0, B: 0, A: 255}
}

// formatCompareValue は reflect.Value を比較用に文字列化します。
func formatCompareValue(v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', 6, 64)
	case reflect.String:
		return v.String()
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
