package log

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

var (
	debugFlag   bool   = false
	debugPrefix string = "[DEBUG]"
	infoPrefix  string = color.HiBlueString("[INFO]")
	warnPrefix  string = color.YellowString("[WARN]")
	errorPrefix string = color.RedString("[ERROR]")
	fatalPrefix string = color.HiRedString("[FATAL]")
)

func Init(dbg bool) {
	log.SetFlags(0)
	log.SetPrefix("")

	debugFlag = dbg
}

func logPrintf(prefix, fmtString string, params ...any) {
	strItself := fmt.Sprintf(fmtString, params...)
	line := fmt.Sprintf("%s %s", prefix, strItself)

	log.Printf("%s\n", line)
}

func logPrintln(prefix string, params ...any) {
	var strItself string = prefix

	for _, item := range params {
		strItself += fmt.Sprintf(" %v", item)
	}

	log.Println(strItself)
}

func Debug(message string, params ...any) {
	if debugFlag {
		logPrintf(debugPrefix, message, params...)
	}
}

func Debugln(params ...any) {
	if debugFlag {
		logPrintln(debugPrefix, params...)
	}
}

func Info(message string, params ...any) {
	logPrintf(infoPrefix, message, params...)
}

func Infoln(params ...any) {
	logPrintln(infoPrefix, params...)
}

func Warn(message string, params ...any) {
	logPrintf(warnPrefix, message, params...)
}

func Warnln(params ...any) {
	logPrintln(warnPrefix, params...)
}

func Error(message string, params ...any) {
	logPrintf(errorPrefix, message, params...)
}

func Errorln(params ...any) {
	logPrintln(debugPrefix, params...)
}

func Fatal(message string, params ...any) {
	logPrintf(fatalPrefix, message, params...)
	os.Exit(1)
}

func Fatalln(params ...any) {
	logPrintln(fatalPrefix, params...)
	os.Exit(1)
}
