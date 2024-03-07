// 🌈 Rainbow logger 🌈
// Replace fmt with rblog to get a timestamp and filename:linenumber prefix in your logs.
// Replace fmt with rblog.Red, rblog.Green, rblog.Yellow, rblog.Blue, rblog.Magenta, or rblog.Cyan to get colored logs.
// Print and Println are equivalent and will both add a newline to the end of the log, but are both kept for drop-in compatibility with fmt.
package rblog

import (
	"fmt"
	"log"
	"os"
)

type color string

var (
	Red     color  = "\033[31m"
	Green   color  = "\033[32m"
	Yellow  color  = "\033[33m"
	Blue    color  = "\033[34m"
	Magenta color  = "\033[35m"
	Cyan    color  = "\033[36m"
	White   color  = "\033[37m"
	reset   string = "\033[0m"
)

var std = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)

func Print(v ...any) {
	std.Output(2, fmt.Sprint(v...))
}

func Printf(format string, v ...any) {
	std.Output(2, fmt.Sprintf(format, v...))
}

func Println(v ...any) {
	std.Output(2, fmt.Sprintln(v...))
}

func (l *color) Print(v ...any) {
	out := string(*l) + fmt.Sprint(v...) + reset
	std.Output(2, out)
}

func (l *color) Printf(format string, v ...any) {
	out := string(*l) + fmt.Sprintf(format, v...) + reset
	std.Output(2, out)
}

func (l *color) Println(v ...any) {
	out := string(*l) + fmt.Sprint(v...) + reset
	std.Output(2, out)
}
