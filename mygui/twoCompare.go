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

// OpenTwoCompare は 2 つのテーブルを比較するためのウィンドウを開きます。
// 指示書に従い、全データを表形式で表示し、異なる値はハイライトします。
// 大きな配列は遅延レンダリング(createLazyGrid)で生成します。
func OpenTwoCompare(app fyne.App) {
	w := app.NewWindow("TwoCompare - Prototype")
	w.Resize(fyne.NewSize(1000, 700))

	// インデックス文字列準備
	numTables := len(mydata.TableList)
	if numTables == 0 {
		w.SetContent(container.NewVBox(widget.NewLabel("No tables loaded")))
		w.Show()
		return
	}
	idxs := make([]string, numTables)
	for i := 0; i < numTables; i++ {
		idxs[i] = fmt.Sprintf("%d", i)
	}

	// UI パーツ
	leftSelect := widget.NewSelect(idxs, nil)
	rightSelect := widget.NewSelect(idxs, nil)
	leftSelect.SetSelectedIndex(0)
	if len(idxs) > 1 {
		rightSelect.SetSelectedIndex(1)
	}

	compareBtn := widget.NewButton("Compare", nil)
	status := widget.NewLabel("")

	// 左右の表示領域（縦スクロールのみ、横幅を固定して横スクロールを防止）
	leftContent := container.NewVBox(widget.NewLabel("Left table not selected"))
	rightContent := container.NewVBox(widget.NewLabel("Right table not selected"))
	leftBox := container.NewVScroll(leftContent)
	rightBox := container.NewVScroll(rightContent)
	// 横幅を固定して、ウィンドウが横に伸びるのを防ぐ
	leftBox.SetMinSize(fyne.NewSize(450, 600))
	rightBox.SetMinSize(fyne.NewSize(450, 600))

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

		// 左右それぞれのカラムを生成してセット（外側の VScroll が縦スクロールを担う）
		leftContent.Objects = []fyne.CanvasObject{buildComparisonColumn(&a, &b, true)}
		leftContent.Refresh()
		rightContent.Objects = []fyne.CanvasObject{buildComparisonColumn(&a, &b, false)}
		rightContent.Refresh()
		status.SetText(fmt.Sprintf("Compared %d ↔ %d", li, ri))
	}

	compareBtn.OnTapped = compareAction

	// コントロール行とレイアウト
	controlRow := container.NewHBox(
		widget.NewLabel("Left:"), leftSelect,
		widget.NewLabel("Right:"), rightSelect,
		compareBtn, status,
	)
	split := container.NewHSplit(container.NewVBox(controlRow, leftBox), rightBox)
	split.SetOffset(0.5)

	w.SetContent(split)
	w.Show()
}

// buildComparisonColumn は比較対象の一方分を縦に並べた表示を返します。
// left==true のとき左側の値を表示し、false のとき右側の値を表示します。
// 両側は同じフィールド順で生成されるため視覚上の整列が取れます。
func buildComparisonColumn(a, b *mydata.TableData, left bool) fyne.CanvasObject {
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
			header = fmt.Sprintf("%s — %s", name, desc)
		}
		// セクション見出し
		sections = append(sections, widget.NewLabelWithStyle(header, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

		fa := va.Field(i)
		fb := vb.Field(i)

		// 配列／スライスは遅延レンダリングのアコーディオンで表示（index + value）
		if fa.Kind() == reflect.Array || fa.Kind() == reflect.Slice {
			title := fmt.Sprintf("%s (array)", name)
			item := widget.NewAccordionItem(title, createLazyGrid(func() fyne.CanvasObject {
				// 表示は index | value(for this side) の 2 列。差分は色で示す（赤=差分）
				// 横幅を制限するため、固定幅のコンテナで囲む
				innerBox := container.NewVBox()
				n := fa.Len()
				for idx := 0; idx < n; idx++ {
					// 値文字列（右側が存在すれば比較して色付け）
					var leftStr, rightStr string
					if idx < fa.Len() {
						leftStr = formatValue(fa.Index(idx))
					} else {
						leftStr = "<nil>"
					}
					if idx < fb.Len() {
						rightStr = formatValue(fb.Index(idx))
					} else {
						rightStr = "<nil>"
					}
					// どちら側を表示するか
					display := leftStr
					other := rightStr
					if !left {
						display = rightStr
						other = leftStr
					}
					c := pickColor(display, other)

					// 1行ずつHBoxで表示（index: value の形式）
					rowLabel := widget.NewLabel(fmt.Sprintf("[%d]:", idx))
					rowLabel.Wrapping = fyne.TextTruncate
					valueText := canvas.NewText(display, c)
					valueText.TextSize = 12
					row := container.NewHBox(rowLabel, valueText)
					innerBox.Add(row)
				}
				return innerBox
			}))
			acc := widget.NewAccordion(item)
			sections = append(sections, acc, widget.NewSeparator())
			continue
		}

		// 構造体は内部フィールドを展開して表示（WeldCode など）
		if fa.Kind() == reflect.Struct {
			innerBox := container.NewVBox()
			// サブフィールドを順に表示（横幅を制限）
			for j := 0; j < fa.NumField(); j++ {
				subName := fa.Type().Field(j).Name
				leftStr := formatValue(fa.Field(j))
				rightStr := formatValue(fb.Field(j))
				display := leftStr
				other := rightStr
				if !left {
					display = rightStr
					other = leftStr
				}

				// フィールド名: 値 の形式で1行ずつ表示
				fieldLabel := widget.NewLabel(fmt.Sprintf("%s:", subName))
				fieldLabel.Wrapping = fyne.TextTruncate
				valueText := canvas.NewText(display, pickColor(display, other))
				valueText.TextSize = 12
				row := container.NewHBox(fieldLabel, valueText)
				innerBox.Add(row)
			}
			sections = append(sections, innerBox, widget.NewSeparator())
			continue
		}

		// プリミティブ値は単一行表示
		leftStr := formatValue(fa)
		rightStr := formatValue(fb)
		display := leftStr
		other := rightStr
		if !left {
			display = rightStr
			other = leftStr
		}
		sections = append(sections, canvas.NewText(display, pickColor(display, other)), widget.NewSeparator())
	}

	return container.NewVBox(sections...)
}

// pickColor は a と b が異なるとき赤、同じなら黒を返します。
func pickColor(a, b string) color.Color {
	if a != b {
		return color.RGBA{R: 200, G: 0, B: 0, A: 255}
	}
	return color.RGBA{R: 0, G: 0, B: 0, A: 255}
}
