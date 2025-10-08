package mygui

import (
	"fmt"

	"NR_Table/mydata"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// OpenDataCompare はデータ集計比較のプロトタイプウィンドウを開きます
func OpenDataCompare(app fyne.App) {
	w := app.NewWindow("DataCompare - Prototype")
	w.Resize(fyne.NewSize(800, 600))

	// プロトタイプ: V番号を選択（シミュレート）してテーブル間のカウントを表示
	vSelect := widget.NewSelect([]string{"V05", "V06", "V32", "V34"}, func(s string) {})
	vSelect.SetSelected("V05")

	info := widget.NewLabel("Select a data key to compare across tables.\nTables loaded: " + fmt.Sprintf("%d", len(mydata.TableList)))

	w.SetContent(container.NewVBox(vSelect, info))
	w.Show()
}
