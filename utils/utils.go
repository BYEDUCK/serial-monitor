package utils

import (
	"container/list"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

const TIME_FORMAT = "15:04:05.000000"

func Must(description string, err error) {
	if err != nil {
		msg := fmt.Sprintf("Cannot %s: %v", description, err)
		log.Fatalln(msg)
	}
}

func MustFn(description string, fn func() error) {
	Must(description, fn())
}

func ListToSliceMsg(l *list.List, maxLen int, printTime bool, printInHex bool) []string {
	var arr []string
	i := 0
	for e := l.Back(); e != nil && i < maxLen; e = e.Prev() {
		msg := e.Value.(*Message)
		var prefix string
		if printTime {
			prefix = fmt.Sprintf("[%s]:", msg.Timestamp.Format(TIME_FORMAT))
		} else {
			prefix = fmt.Sprintf("[%d]:", i+1)
		}
		switch t := msg.Content.(type) {
		case string:
			if printInHex {
				arr = append(arr, append([]string{fmt.Sprintf("%s %s", prefix, strings.TrimRight(msg.Content.(string), "\r\n"))}, toHexLines(msg.Content.(string))...)...)
			} else {
				arr = append(arr, fmt.Sprintf("%s %s", prefix, strings.TrimRight(msg.Content.(string), "\r\n")))
			}
		case float64:
			arr = append(arr, fmt.Sprintf("%s %f", prefix, msg.Content.(float64)))
		default:
			log.Fatalf("Unknown msg type: %v\n", t)
		}
		i++
	}
	return arr[:]
}

func ListToSliceFloat(l *list.List, maxLen int) []float64 {
	arr := make([]float64, maxLen)
	i := maxLen - 1
	for e := l.Front(); e != nil && i >= 0; e = e.Next() {
		msg := e.Value.(*Message)
		switch t := msg.Content.(type) {
		case float64:
			arr[i] = msg.Content.(float64)
		case string:
			num, err := strconv.ParseFloat(strings.TrimRight(msg.Content.(string), "\r\n"), 64)
			if err == nil {
				arr[i] = num
			} else {
				i++
			}
		default:
			log.Fatalf("Unknown msg type: %v\n", t)
		}
		i--
	}
	return arr[i+1:]
}

func toHexLines(s string) []string {
	hexString := strings.ToUpper(hex.EncodeToString([]byte(s)))
	linesCount := int(math.Ceil(float64(len(hexString)) / 16.0))
	hexLines := make([]string, linesCount)
	for i := 0; i < linesCount; i++ {
		start := i * 16
		end := (i + 1) * 16
		if end > len(hexString) {
			end = len(hexString)
		}
		hexLine := hexString[start:end]
		hexLines[i] = splitHexLine(hexLine, i)
	}
	return hexLines[:]
}

func splitHexLine(hexLine string, lineNum int) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%04x", lineNum*8))
	builder.WriteRune('-')
	builder.WriteString(fmt.Sprintf("%04x", lineNum*8+7)) // end address points to last byte
	builder.WriteString("  ")
	for i, r := range hexLine {
		if i != 0 && i%2 == 0 {
			builder.WriteRune(' ')
		}
		builder.WriteRune(r)
	}
	return builder.String()
}
