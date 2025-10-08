package myast

import (
	"NR_Table/mydata"
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParseWeldTable は weldtbl.c を解析して mydata.TableList に格納します
func ParseWeldTable(filepath string) error {
	// Shift_JIS でファイルを開きます
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Shift_JIS デコーダを作成します
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(file, decoder)
	scanner := bufio.NewScanner(reader)

	// 全ての行を読み込みます
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// テーブルをパースします
	tables, err := extractTables(lines)
	if err != nil {
		return fmt.Errorf("failed to extract tables: %w", err)
	}

	// mydata.TableList に格納します
	mydata.TableList = tables

	return nil
}

// extractTables はソース行から WELDTABLE_GX3 のテーブルデータを抽出します
func extractTables(lines []string) ([]mydata.TableData, error) {
	var tables []mydata.TableData

	// テーブル定義の開始位置を探します: "WELDTABLE_GX3 const WeldTable[ ] ="
	tableStart := -1
	for i, line := range lines {
		if strings.Contains(line, "WELDTABLE_GX3") && strings.Contains(line, "WeldTable") {
			tableStart = i
			break
		}
	}

	if tableStart == -1 {
		return nil, fmt.Errorf("WELDTABLE_GX3 declaration not found")
	}

	// 解析用の状態マシン
	type ParseState int
	const (
		StateSearchTable ParseState = iota
		StateInWeldCode
		StateInA2SPulse
		StateInS2VPulse
		StateInA2SShort
		StateInS2VShort
		StateInWeldParm
		StateInCalParm
		StateOther
	)

	state := StateSearchTable
	currentTable := mydata.TableData{}
	braceDepth := 0
	var currentArray []int
	arrayIndex := 0
	weldCodeValues := []uint8{} // WeldCode の値を蓄積します

	for i := tableStart; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// コメント行とプリプロセッサディレクティブをスキップします
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 波括弧の深さをカウントします
		for _, ch := range line {
			if ch == '{' {
				braceDepth++
			} else if ch == '}' {
				braceDepth--
			}
		}

		// コメントに基づく状態遷移
		if strings.Contains(trimmed, "{// 溶接種別コード") || strings.Contains(trimmed, "{//") && state == StateSearchTable {
			state = StateInWeldCode
			arrayIndex = 0
			weldCodeValues = []uint8{} // Reset WeldCode values
			continue
		}

		// コメントによりセクション変更を検出します
		if strings.Contains(trimmed, "// パルス溶接用ワイヤ送給速度テーブル") {
			// Save accumulated WeldCode values
			if len(weldCodeValues) >= 12 {
				currentTable.WeldCode.Material = weldCodeValues[0]
				currentTable.WeldCode.Method = weldCodeValues[1]
				currentTable.WeldCode.PulseMode = weldCodeValues[2]
				currentTable.WeldCode.PulseType = weldCodeValues[3]
				currentTable.WeldCode.Wire = weldCodeValues[4]
				currentTable.WeldCode.Extension = weldCodeValues[5]
				currentTable.WeldCode.Tip = weldCodeValues[6]
				currentTable.WeldCode.Flag2 = weldCodeValues[7]
				currentTable.WeldCode.Version = weldCodeValues[8]
				currentTable.WeldCode.StandardFlag = weldCodeValues[9]
				currentTable.WeldCode.Flag3 = weldCodeValues[10]
				currentTable.WeldCode.LowSputter = weldCodeValues[11]
			}
			state = StateInA2SPulse
			arrayIndex = 0
			currentArray = []int{}
			continue
		} else if strings.Contains(trimmed, "// パルス溶接用一元電圧テーブル") {
			if state == StateInA2SPulse && len(currentArray) > 0 {
				copyToA2S(&currentTable.A2S_Pulse, currentArray)
			}
			state = StateInS2VPulse
			arrayIndex = 0
			currentArray = []int{}
			continue
		} else if strings.Contains(trimmed, "// 短絡溶接用ワイヤ送給速度テーブル") {
			if state == StateInS2VPulse && len(currentArray) > 0 {
				copyToS2V(&currentTable.S2V_Pulse, currentArray)
			}
			state = StateInA2SShort
			arrayIndex = 0
			currentArray = []int{}
			continue
		} else if strings.Contains(trimmed, "// 短絡溶接用一元電圧テーブル") {
			if state == StateInA2SShort && len(currentArray) > 0 {
				copyToA2S(&currentTable.A2S_Short, currentArray)
			}
			state = StateInS2VShort
			arrayIndex = 0
			currentArray = []int{}
			continue
		} else if strings.Contains(trimmed, "// 半固定パラメータテーブル") {
			if state == StateInS2VShort && len(currentArray) > 0 {
				copyToS2V(&currentTable.S2V_Short, currentArray)
			}
			state = StateInWeldParm
			arrayIndex = 0
			currentArray = []int{}
			continue
		} else if strings.Contains(trimmed, "// 可変パラーメータ係数テーブル") || strings.Contains(trimmed, "// 可変パラメータ係数テーブル") {
			if state == StateInWeldParm && len(currentArray) > 0 {
				copyToWeldParm(&currentTable.WeldParm, currentArray)
			}
			state = StateInCalParm
			arrayIndex = 0
			continue
		}

		// 現在の状態に基づいてデータをパースします
		switch state {
		case StateInWeldCode:
			// Extract hex values and accumulate
			hexPattern := regexp.MustCompile(`0x([0-9A-Fa-f]{2})`)
			matches := hexPattern.FindAllStringSubmatch(trimmed, -1)
			for _, match := range matches {
				if len(match) > 1 {
					val, _ := strconv.ParseUint(match[1], 16, 8)
					weldCodeValues = append(weldCodeValues, uint8(val))
				}
			}

		case StateInA2SPulse, StateInS2VPulse, StateInA2SShort, StateInS2VShort, StateInWeldParm:
			values := extractNumbers(trimmed)
			currentArray = append(currentArray, values...)

		case StateInCalParm:
			parseCalParm(&currentTable, trimmed, &arrayIndex)
		}

		// テーブルエントリの終わりかを判定します
		if braceDepth == 1 && strings.Contains(trimmed, "},") {
			// Save accumulated arrays
			switch state {
			case StateInWeldCode:
				// Save WeldCode if still in that state
				if len(weldCodeValues) >= 12 {
					currentTable.WeldCode.Material = weldCodeValues[0]
					currentTable.WeldCode.Method = weldCodeValues[1]
					currentTable.WeldCode.PulseMode = weldCodeValues[2]
					currentTable.WeldCode.PulseType = weldCodeValues[3]
					currentTable.WeldCode.Wire = weldCodeValues[4]
					currentTable.WeldCode.Extension = weldCodeValues[5]
					currentTable.WeldCode.Tip = weldCodeValues[6]
					currentTable.WeldCode.Flag2 = weldCodeValues[7]
					currentTable.WeldCode.Version = weldCodeValues[8]
					currentTable.WeldCode.StandardFlag = weldCodeValues[9]
					currentTable.WeldCode.Flag3 = weldCodeValues[10]
					currentTable.WeldCode.LowSputter = weldCodeValues[11]
				}
			case StateInA2SPulse:
				copyToA2S(&currentTable.A2S_Pulse, currentArray)
			case StateInS2VPulse:
				copyToS2V(&currentTable.S2V_Pulse, currentArray)
			case StateInA2SShort:
				copyToA2S(&currentTable.A2S_Short, currentArray)
			case StateInS2VShort:
				copyToS2V(&currentTable.S2V_Short, currentArray)
			case StateInWeldParm:
				copyToWeldParm(&currentTable.WeldParm, currentArray)
			}

			tables = append(tables, currentTable)
			currentTable = mydata.TableData{}
			state = StateSearchTable
			currentArray = []int{}
			arrayIndex = 0
			weldCodeValues = []uint8{}
		}

		// 配列全体の終端かを判定します
		if braceDepth == 0 && strings.Contains(trimmed, "};") {
			break
		}
	}

	return tables, nil
}

// extractNumbers は行から数値をすべて抽出して返します
func extractNumbers(line string) []int {
	numPattern := regexp.MustCompile(`[-]?(?:0x[0-9A-Fa-f]+|\d+)`)
	matches := numPattern.FindAllString(line, -1)

	var values []int
	for _, match := range matches {
		var val int64
		if strings.HasPrefix(match, "0x") || strings.HasPrefix(match, "0X") {
			val, _ = strconv.ParseInt(match[2:], 16, 32)
		} else {
			val, _ = strconv.ParseInt(match, 10, 32)
		}
		values = append(values, int(val))
	}
	return values
}

// copyToA2S は整数配列を A2STBL にコピーします
func copyToA2S(dest *mydata.A2STBL, src []int) {
	for i := 0; i < len(src) && i < 256; i++ {
		dest.Speed[i] = uint16(src[i])
	}
}

// copyToS2V は整数配列を S2VTBL にコピーします
func copyToS2V(dest *mydata.S2VTBL, src []int) {
	for i := 0; i < len(src) && i < 256; i++ {
		dest.Values[i] = int16(src[i])
	}
}

// copyToWeldParm は整数配列を WELDPARM にコピーします
func copyToWeldParm(dest *mydata.WELDPARM, src []int) {
	for i := 0; i < len(src) && i < 504; i++ {
		dest.Parm[i] = uint16(src[i])
	}
}

// parseCalParm は可変パラメータ係数セクションをパースします
func parseCalParm(table *mydata.TableData, line string, index *int) {
	// {0,1,0,0,3} のようなパターンを探します
	if strings.Contains(line, "{") && strings.Contains(line, "}") {
		// Extract floats between braces
		start := strings.Index(line, "{")
		end := strings.Index(line, "}")
		if start >= 0 && end > start {
			content := line[start+1 : end]
			parts := strings.Split(content, ",")

			if len(parts) == 5 && *index < 116 {
				// This is a CalParm entry (stored as 5 consecutive floats)
				// For simplicity, just store the first value
				val, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 32)
				if err == nil && *index < 116 {
					table.CalParm[*index] = float32(val)
					*index++
				}
			}
		}
	}
}
