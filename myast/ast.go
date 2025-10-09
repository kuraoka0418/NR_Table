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
		StateInCalParmList
		StateInVxxData
	)

	state := StateSearchTable
	currentTable := mydata.TableData{}
	braceDepth := 0
	var currentArray []int
	var currentFloatArray []float32
	weldCodeValues := []uint8{}
	currentVxxIndex := -1 // 現在パース中のVxxデータのインデックス

	for i := tableStart; i < len(lines); i++ {
		originalLine := lines[i] // コメント除去前の元の行を保存（セクション判定に使用）

		// コメント除去と空白文字の正規化を行います
		line := removeLineComment(originalLine)
		line = normalizeWhitespace(line)
		trimmed := line

		// 空行やプリプロセッサディレクティブをスキップします
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
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

		// テーブルエントリの開始を検出します
		if braceDepth == 1 && strings.Contains(trimmed, "{") && state == StateSearchTable {
			state = StateInWeldCode
			weldCodeValues = []uint8{}
			continue
		}

		// コメントによりセクション変更を検出します
		// セクション判定には元の行（コメント除去前）を使用します

		if strings.Contains(originalLine, "パルス溶接用ワイヤ送給速度テーブル") {
			if len(weldCodeValues) >= 12 {
				assignWeldCode(&currentTable, weldCodeValues)
			}
			state = StateInA2SPulse
			currentArray = []int{}
			continue
		} else if strings.Contains(originalLine, "パルス溶接用一元電圧テーブル") {
			if state == StateInA2SPulse && len(currentArray) > 0 {
				copyToA2S(&currentTable.A2S_Pulse, currentArray)
			}
			state = StateInS2VPulse
			currentArray = []int{}
			continue
		} else if strings.Contains(originalLine, "短絡溶接用ワイヤ送給速度テーブル") {
			if state == StateInS2VPulse && len(currentArray) > 0 {
				copyToS2V(&currentTable.S2V_Pulse, currentArray)
			}
			state = StateInA2SShort
			currentArray = []int{}
			continue
		} else if strings.Contains(originalLine, "短絡溶接用一元電圧テーブル") {
			if state == StateInA2SShort && len(currentArray) > 0 {
				copyToA2S(&currentTable.A2S_Short, currentArray)
			}
			state = StateInS2VShort
			currentArray = []int{}
			continue
		} else if strings.Contains(originalLine, "半固定パラメータテーブル") {
			if state == StateInS2VShort && len(currentArray) > 0 {
				copyToS2V(&currentTable.S2V_Short, currentArray)
			}
			state = StateInWeldParm
			currentArray = []int{}
			continue
		} else if strings.Contains(originalLine, "可変パラ") && strings.Contains(originalLine, "メータ係数テーブル") {
			// "可変パラーメータ係数テーブル" を検出
			if state == StateInWeldParm && len(currentArray) > 0 {
				copyToWeldParm(&currentTable.WeldParm, currentArray)
			}
			state = StateInCalParm
			currentFloatArray = []float32{}
			continue
		} else if strings.Contains(originalLine, "パラメータテーブル（パルス）") ||
			strings.Contains(originalLine, "パラメータテーブル（短絡）") {
			// CalParm終了を検出
			if state == StateInCalParm && len(currentFloatArray) > 0 {
				copyToCalParm(&currentTable, currentFloatArray)
				currentFloatArray = []float32{}
			}
			state = StateInCalParmList
			continue
		}

		// Vxxデータの開始を検出（コメント内のV5, V6, V8...など）
		if strings.Contains(originalLine, "{//V") || strings.Contains(originalLine, "{// V") {
			// 前のVxxデータを保存
			if currentVxxIndex >= 0 && len(currentArray) > 0 {
				copyToVxxData(&currentTable, currentVxxIndex, currentArray)
			}

			// Vxx番号を抽出
			vxxPattern := regexp.MustCompile(`V(\d+)`)
			if matches := vxxPattern.FindStringSubmatch(originalLine); len(matches) > 1 {
				if vnum, err := strconv.Atoi(matches[1]); err == nil {
					currentVxxIndex = vnum
					currentArray = []int{}
					state = StateInVxxData
				}
			}
		}

		// 現在の状態に基づいてデータをパースします
		switch state {
		case StateInWeldCode:
			// 16進数値を抽出して蓄積します
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
			// CalParm は {a, b, c, min, max} の形式で5つの float 値を持ちます
			// 複数行にまたがる可能性があるため、{ と } を探して抽出します
			if strings.Contains(trimmed, "{") {
				start := strings.Index(trimmed, "{")
				end := strings.Index(trimmed, "}")
				if end < 0 {
					// 次の行に続く可能性がある
					content := trimmed[start+1:]
					floats := extractFloats(content)
					currentFloatArray = append(currentFloatArray, floats...)
				} else {
					content := trimmed[start+1 : end]
					floats := extractFloats(content)
					currentFloatArray = append(currentFloatArray, floats...)
				}
			} else if strings.Contains(trimmed, "}") {
				// 前の行から続いている場合
				end := strings.Index(trimmed, "}")
				content := trimmed[:end]
				floats := extractFloats(content)
				currentFloatArray = append(currentFloatArray, floats...)
			} else if len(trimmed) > 0 && !strings.Contains(trimmed, "CalParm") {
				// 途中の行
				floats := extractFloats(trimmed)
				currentFloatArray = append(currentFloatArray, floats...)
			}

		case StateInCalParmList:
			// CalParmList は { V番号, 係数 } の形式
			// この部分は現在のデータ構造では保存していないため、スキップします
			continue

		case StateInVxxData:
			// Vxxデータは整数配列
			if currentVxxIndex >= 0 {
				values := extractNumbers(trimmed)
				currentArray = append(currentArray, values...)
			}
		}

		// テーブルエントリの終わりを検出します (braceDepth が 1 に戻り、"}, " がある場合)
		if braceDepth == 1 && strings.Contains(trimmed, "},") {
			// 最後のVxxデータを保存
			if state == StateInVxxData && currentVxxIndex >= 0 && len(currentArray) > 0 {
				copyToVxxData(&currentTable, currentVxxIndex, currentArray)
			}

			// 蓄積した配列を保存します
			switch state {
			case StateInWeldCode:
				if len(weldCodeValues) >= 12 {
					assignWeldCode(&currentTable, weldCodeValues)
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
			case StateInCalParm:
				copyToCalParm(&currentTable, currentFloatArray)
			}

			tables = append(tables, currentTable)
			currentTable = mydata.TableData{}
			state = StateSearchTable
			currentArray = []int{}
			currentFloatArray = []float32{}
			weldCodeValues = []uint8{}
			currentVxxIndex = -1
		}

		// 配列全体の終端を判定します
		if braceDepth == 0 && strings.Contains(trimmed, "};") {
			break
		}
	}

	return tables, nil
}

// extractNumbers は行から数値をすべて抽出して返します (整数のみ、16進数と10進数に対応)
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

// extractFloats は行から浮動小数点数をすべて抽出して返します
func extractFloats(line string) []float32 {
	// 浮動小数点数パターン: 整数部.小数部、整数のみ、負の数に対応
	floatPattern := regexp.MustCompile(`[-]?\d+\.?\d*`)
	matches := floatPattern.FindAllString(line, -1)

	var values []float32
	for _, match := range matches {
		match = strings.TrimSpace(match)
		if match == "" || match == "-" {
			continue
		}
		val, err := strconv.ParseFloat(match, 32)
		if err == nil {
			values = append(values, float32(val))
		}
	}
	return values
}

// removeLineComment は行から // コメントを除去します（文字列リテラル外のみ）
func removeLineComment(line string) string {
	// シンプルに // の位置を探して、それ以降を削除します
	// 文字列リテラル内の "//" は C コードでは稀なので、単純な実装で十分
	if idx := strings.Index(line, "//"); idx >= 0 {
		return line[:idx]
	}
	return line
}

// normalizeWhitespace は行の空白文字を正規化します
func normalizeWhitespace(line string) string {
	// タブをスペースに変換
	line = strings.ReplaceAll(line, "\t", " ")
	// 複数の連続するスペースを1つに圧縮
	line = regexp.MustCompile(`\s+`).ReplaceAllString(line, " ")
	return strings.TrimSpace(line)
}

// assignWeldCode は WeldCode 構造体に値を割り当てます
func assignWeldCode(table *mydata.TableData, values []uint8) {
	if len(values) >= 12 {
		table.WeldCode.Material = values[0]
		table.WeldCode.Method = values[1]
		table.WeldCode.PulseMode = values[2]
		table.WeldCode.PulseType = values[3]
		table.WeldCode.Wire = values[4]
		table.WeldCode.Extension = values[5]
		table.WeldCode.Tip = values[6]
		table.WeldCode.Flag2 = values[7]
		table.WeldCode.Version = values[8]
		table.WeldCode.StandardFlag = values[9]
		table.WeldCode.Flag3 = values[10]
		table.WeldCode.LowSputter = values[11]
	}
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

// copyToCalParm は浮動小数点配列を CalParm にコピーします
func copyToCalParm(table *mydata.TableData, src []float32) {
	// CalParm は [116]float32 配列
	for i := 0; i < len(src) && i < 116; i++ {
		table.CalParm[i] = src[i]
	}
}

// copyToVxxData は整数配列を対応する Vxx データにコピーします
func copyToVxxData(table *mydata.TableData, vnum int, src []int) {
	// Vxx データは [128]int16 配列
	copyInt16 := func(dest *[128]int16) {
		for i := 0; i < len(src) && i < 128; i++ {
			dest[i] = int16(src[i])
		}
	}

	switch vnum {
	case 5:
		copyInt16(&table.V05_Data)
	case 6:
		copyInt16(&table.V06_Data)
	case 8:
		copyInt16(&table.V08_Data)
	case 12:
		copyInt16(&table.V12_Data)
	case 13:
		copyInt16(&table.V13_Data)
	case 15:
		copyInt16(&table.V15_Data)
	case 18:
		copyInt16(&table.V18_Data)
	case 19:
		copyInt16(&table.V19_Data)
	case 20:
		copyInt16(&table.V20_Data)
	case 32:
		copyInt16(&table.V32_Data)
	case 34:
		copyInt16(&table.V34_Data)
	case 36:
		copyInt16(&table.V36_Data)
	case 56:
		copyInt16(&table.V56_Data)
	case 57:
		copyInt16(&table.V57_Data)
	case 59:
		copyInt16(&table.V59_Data)
	case 68:
		copyInt16(&table.V68_Data)
	case 93:
		copyInt16(&table.V93_Data)
	case 94:
		copyInt16(&table.V94_Data)
	case 95:
		copyInt16(&table.V95_Data)
	}
}
