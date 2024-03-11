// 🌈 Rainbow logger 🌈
// Replace fmt with rblog to get a timestamp and filename:linenumber prefix in your logs.
// Replace fmt with rblog.Red, rblog.Green, rblog.Yellow, rblog.Blue, rblog.Magenta, or rblog.Cyan to get colored logs.
// Replace fmt with rblog.Rainbow to get rainbow logs🌈🌈🌈.
// Print and Println are equivalent and will both add a newline to the end of the log, but are both kept for drop-in compatibility with fmt.
package rblog

import (
	"fmt"
	"log"
	"os"
	"strings"
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

type specialColor string

var Rainbow specialColor = "🌈"

var rainbowSequence = []string{"\033[1;37;41m", "\033[1;30;43m", "\033[1;30;42m", "\033[1;30;46m", "\033[1;37;44m", "\033[1;37;45m"}

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

func (c *color) Print(v ...any) {
	s := fmt.Sprint(v...)
	out := string(*c) + strings.TrimSuffix(s, "\n") + reset
	std.Output(2, out)
}

func (c *color) Printf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	out := string(*c) + strings.TrimSuffix(s, "\n") + reset
	std.Output(2, out)
}

func (c *color) Println(v ...any) {
	out := string(*c) + fmt.Sprint(v...) + reset
	std.Output(2, out)
}

func (c *specialColor) Print(v ...any) {
	s := rainbowify(fmt.Sprint(v...))
	std.Output(2, s)
}

func (c *specialColor) Printf(format string, v ...any) {
	s := rainbowify(fmt.Sprintf(format, v...))
	std.Output(2, s)
}

func (c *specialColor) Println(v ...any) {
	s := rainbowify(fmt.Sprint(v...))
	std.Output(2, s)
}

func rainbowify(s string) string {
	result := ""
	for i, char := range s {
		result += rainbowSequence[i%len(rainbowSequence)] + string(char)
	}
	// Remove the newline at the end of string to avoid double newlines
	// Newline will be added back by std.Output

	return strings.TrimSuffix(result, "\n") + reset
}
