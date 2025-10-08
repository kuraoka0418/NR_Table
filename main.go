package main

import (
	"NR_Table/myast"
	"NR_Table/mydata"
	"NR_Table/mygui"
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2/app"
)

func main() {
	// Load table data from weldtbl.c
	weldtblPath := filepath.Join("res", "weldtbl.c")
	err := myast.ParseWeldTable(weldtblPath)
	if err != nil {
		fmt.Printf("Warning: Failed to parse weldtbl.c: %v\n", err)
		fmt.Println("Loading dummy data instead...")
		mydata.InitDummyData()
	} else {
		fmt.Printf("Successfully loaded %d tables from %s\n", mydata.GetTableCount(), weldtblPath)

		// Debug: Show first table data
		if len(mydata.TableList) > 0 {
			t := mydata.TableList[0]
			fmt.Printf("\nFirst table sample:\n")
			fmt.Printf("  WeldCode.Material: %d\n", t.WeldCode.Material)
			fmt.Printf("  WeldCode.Method: %d\n", t.WeldCode.Method)
			fmt.Printf("  A2S_Short.Speed[0-4]: [%d, %d, %d, %d, %d]\n",
				t.A2S_Short.Speed[0], t.A2S_Short.Speed[1], t.A2S_Short.Speed[2],
				t.A2S_Short.Speed[3], t.A2S_Short.Speed[4])
			fmt.Printf("  S2V_Short.Values[0-4]: [%d, %d, %d, %d, %d]\n",
				t.S2V_Short.Values[0], t.S2V_Short.Values[1], t.S2V_Short.Values[2],
				t.S2V_Short.Values[3], t.S2V_Short.Values[4])
			fmt.Printf("  WeldParm.Parm[0-4]: [0x%04X, 0x%04X, 0x%04X, 0x%04X, 0x%04X]\n",
				t.WeldParm.Parm[0], t.WeldParm.Parm[1], t.WeldParm.Parm[2],
				t.WeldParm.Parm[3], t.WeldParm.Parm[4])
			fmt.Printf("  CalParm[0-4]: [%.2f, %.2f, %.2f, %.2f, %.2f]\n\n",
				t.CalParm[0], t.CalParm[1], t.CalParm[2], t.CalParm[3], t.CalParm[4])
		}
	}

	// Start GUI
	a := app.New()
	mw := mygui.NewMw(a)
	mw.ShowAndRun()
}
