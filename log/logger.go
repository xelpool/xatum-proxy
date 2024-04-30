package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var LogLevel uint8 = 0

var Stdout io.Writer = os.Stdout
var Stderr io.Writer = os.Stderr

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"
var Bold = "\033[1m"

func getLogPrefix() string {
	_, file, line, _ := runtime.Caller(2)
	fileSpl := strings.Split(file, "/")
	debugInfos := strings.Split(fileSpl[len(fileSpl)-1], ".")[0] + ":" + strconv.FormatInt(int64(line), 10)
	for len(debugInfos) < 16 {
		debugInfos = debugInfos + " "
	}

	return debugInfos
}
func getMutPefix() string {
	_, file, line, _ := runtime.Caller(3)
	fileSpl := strings.Split(file, "/")
	debugInfos := strings.Split(fileSpl[len(fileSpl)-1], ".")[0] + ":" + strconv.FormatInt(int64(line), 10)
	for len(debugInfos) < 16 {
		debugInfos = debugInfos + " "
	}

	return debugInfos
}

func Title(a ...any) {
	Stdout.Write([]byte(fmt.Sprintln(a...) + Reset))
}

func Info(a ...any) {
	Stdout.Write([]byte(getLogPrefix() + "[INFO]  " + fmt.Sprintln(a...) + Reset))
}
func Infof(format string, a ...any) {
	Stdout.Write([]byte(getLogPrefix() + fmt.Sprintf("[INFO]  "+format+"\n", a...) + Reset))
}

func Warn(a ...any) {
	Stdout.Write([]byte(getLogPrefix() + Yellow + "[WARN]  " + fmt.Sprintln(a...) + Reset))
}
func Warnf(format string, a ...any) {
	Stdout.Write([]byte(getLogPrefix() + Yellow + fmt.Sprintf("[WARN]  "+format+"\n", a...) + Reset))
}

func Err(a ...any) {
	Stdout.Write([]byte(getLogPrefix() + Red + "[ERR]   " + fmt.Sprintln(a...) + Reset))
}

func Errf(format string, a ...any) {
	Stderr.Write([]byte(getLogPrefix() + Red + fmt.Sprintf("[ERR]   "+format+"\n", a...) + Reset))
}

func Debug(a ...any) {
	if LogLevel < 1 {
		return
	}

	Stdout.Write([]byte(getLogPrefix() + Cyan + "[DEBUG] " + fmt.Sprintln(a...) + Reset))
}
func Debugf(format string, a ...any) {
	if LogLevel < 1 {
		return
	}

	Stdout.Write([]byte(getLogPrefix() + Cyan + fmt.Sprintf("[DEBUG] "+format+"\n", a...) + Reset))
}

func Dev(a ...any) {
	if LogLevel < 2 {
		return
	}

	Stdout.Write([]byte(getLogPrefix() + Cyan + "[DEV]   " + fmt.Sprintln(a...) + Reset))
}

func Devf(format string, a ...any) {
	if LogLevel < 2 {
		return
	}

	Stdout.Write([]byte(getLogPrefix() + Cyan + fmt.Sprintf("[DEV]   "+format+"\n", a...) + Reset))
}

func DEBUG(a ...any) {
	Stdout.Write([]byte(getLogPrefix() + Red + Bold + "[DEBUG] " + fmt.Sprintln(a...) + Reset))
}
func DEBUGF(format string, a ...any) {
	Stdout.Write([]byte(getLogPrefix() + Red + Bold + fmt.Sprintf("[DEBUG] "+format+"\n", a...) + Reset))
}

func Mutex(format string, a ...any) {
	if LogLevel < 3 {
		return
	}

	Stdout.Write([]byte(getMutPefix() + Purple + fmt.Sprintf("[MUTEX] "+format+"\n", a...) + Reset))
}

func Net(a ...any) {
	if LogLevel < 1 {
		return
	}
	Stdout.Write([]byte(getLogPrefix() + Green + "[NET]   " + fmt.Sprintln(a...) + Reset))
}
func Netf(format string, a ...any) {
	if LogLevel < 1 {
		return
	}

	Stdout.Write([]byte(getLogPrefix() + Green + fmt.Sprintf("[NET]   "+format+"\n", a...) + Reset))
}

func NetDev(a ...any) {
	if LogLevel < 1 {
		return
	}
	Stdout.Write([]byte(getLogPrefix() + Green + "NETDEV  " + fmt.Sprintln(a...) + Reset))
}

func NetDevf(format string, a ...any) {
	if LogLevel < 1 {
		return
	}

	Stdout.Write([]byte(getLogPrefix() + Green + fmt.Sprintf("NETDEV  "+format+"\n", a...) + Reset))
}

func Fatal(err any) {
	Stderr.Write([]byte(getLogPrefix() + Red + fmt.Sprintln("[FATAL]", err) + Reset))
	panic(err)
}
