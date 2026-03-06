package logx

import (
	"encoding/json"
	"log"
)

const gap = " "

func Fatal(v ...any) {
	xs := append([]any{"❌"}, v...)
	log.Fatal(xs...)
}

func Fatalf(format string, v ...any) {
	log.Fatalf("❌ "+format, v...)
}

func Error(v ...any) {
	xs := append([]any{"❌"}, v...)
	log.Print(xs...)
}

func Errorf(format string, v ...any) {
	log.Printf("❌ "+format, v...)
}

func Warn(v ...any) { //width: A not W
	xs := append([]any{"⚠️", gap}, v...)
	log.Print(xs...)
}

func Warnf(format string, v ...any) {
	log.Printf("⚠️ "+gap+format, v...)
}

func Pin(v ...any) {
	xs := append([]any{"📌"}, v...)
	log.Println(xs...)
}

func Pinf(format string, v ...any) {
	log.Printf("📌 "+format, v...)
}

func PinJSON(v ...any) {
	xs := []any{"📌"}
	for _, x := range v {
		str, ok := x.(string)
		if ok {
			xs = append(xs, str)
			continue
		}
		j, _ := json.MarshalIndent(x, "", "  ")
		xs = append(xs, string(j))
	}
	log.Println(xs...)
}

func Debug(v ...any) {
	xs := append([]any{"🐛"}, v...)
	log.Println(xs...)
}

func Debugf(format string, v ...any) {
	log.Printf("🐛 "+format, v...)
}

func DebugJSON(v ...any) {
	xs := []any{"🐛"}
	for _, x := range v {
		str, ok := x.(string)
		if ok {
			xs = append(xs, str)
			continue
		}
		j, _ := json.MarshalIndent(x, "", "  ")
		xs = append(xs, string(j))
	}
	log.Println(xs...)
}
