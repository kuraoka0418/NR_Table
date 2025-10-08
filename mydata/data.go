package mydata

const PRESET_CALPARM_LIST_MAX = 16

type DCCALPARM struct {
	A   float32
	B   float32
	C   float32
	Min float32
	Max float32
}

type WELDPARM struct {
	Parm [504]uint16
}

type CALPARMDATATBL struct {
	Data [128]int16
}

type A2STBL struct {
	Speed [256]uint16
}

type S2VTBL struct {
	Values [256]int16
}

type WELDCODE struct {
	Material     uint8
	Method       uint8
	PulseMode    uint8
	PulseType    uint8
	Wire         uint8
	Extension    uint8
	Tip          uint8
	Flag2        uint8
	Version      uint8
	StandardFlag uint8
	Flag3        uint8
	LowSputter   uint8
}

type TableData struct {
	WeldCode         WELDCODE
	A2S_Pulse        A2STBL
	S2V_Pulse        S2VTBL
	A2S_Short        A2STBL
	S2V_Short        S2VTBL
	WeldParm         WELDPARM
	CalParm          [116]float32
	ParmTbl_Pls      []int16
	ParmTbl_Short    []int16
	CalParmList      []int16
	V05_Data         [128]int16
	V06_Data         [128]int16
	V08_Data         [128]int16
	V12_Data         [128]int16
	V32_Data         [128]int16
	V34_Data         [128]int16
	V36_Data         [128]int16
	V56_Data         [128]int16
	V59_Data         [128]int16
	V68_Data         [128]int16
	V13_Data         [128]int16
	V15_Data         [128]int16
	V18_Data         [128]int16
	V19_Data         [128]int16
	V20_Data         [128]int16
	V94_Data         [128]int16
	V95_Data         [128]int16
	V57_Data         [128]int16
	V93_Data         [128]int16
	CalParmDataTable [PRESET_CALPARM_LIST_MAX - 1]CALPARMDATATBL
	Navi_Pram1       [7]float32
	Navi_Pram2       [7]float32
	Navi_Pram3       [7]float32
	Navi_P_Pram1     [7]float32
	Navi_P_Pram2     [7]float32
	Navi_P_Pram3     [7]float32
}

// テーブルのリスト
var TableList []TableData

// InitDummyData はテスト用の最小限のダミーテーブルを追加します
func InitDummyData() {
	// Add two simple dummy tables
	t1 := TableData{}
	t1.WeldCode.Material = 1
	t1.WeldCode.Method = 2

	t2 := TableData{}
	t2.WeldCode.Material = 10
	t2.WeldCode.Method = 20

	TableList = append(TableList, t1, t2)
}

// GetTableCount は、読み込まれているテーブルの数を返します
func GetTableCount() int {
	return len(TableList)
}
