package mygui

import (
	"fmt"

	"NR_Table/mydata"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// OpenTwoCompare は 2 つのテーブルを比較するための試作ウィンドウを開きます
func OpenTwoCompare(app fyne.App) {
	w := app.NewWindow("TwoCompare - Prototype")
	w.Resize(fyne.NewSize(800, 600))

	// 簡易プロトタイプUI：2つのドロップダウンでテーブルを選び、左右に並べて表示します
	numTables := len(mydata.TableList)
	idxs := make([]string, numTables)
	for i := 0; i < numTables; i++ {
		idxs[i] = fmt.Sprintf("%d", i)
	}

	leftSelect := widget.NewSelect(idxs, func(s string) {})
	rightSelect := widget.NewSelect(idxs, func(s string) {})

	leftLabel := widget.NewLabel("Left table")
	rightLabel := widget.NewLabel("Right table")

	if len(idxs) > 0 {
		leftSelect.SetSelectedIndex(0)
	}
	if len(idxs) > 1 {
		rightSelect.SetSelectedIndex(1)
	}

	content := container.NewHSplit(
		container.NewVBox(leftSelect, leftLabel),
		container.NewVBox(rightSelect, rightLabel),
	)
	content.SetOffset(0.5)

	w.SetContent(content)
	w.Show()
}
