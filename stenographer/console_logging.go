package stenographer

import (
	"fmt"
	"strings"
)

func (s *Stenographer) colorize(colorCode string, format string, args ...interface{}) string {
	var out string

	if len(args) > 0 {
		out = fmt.Sprintf(format, args...)
	} else {
		out = format
	}

	if s.color {
		return fmt.Sprintf("%s%s%s", colorCode, out, defaultStyle)
	} else {
		return out
	}
}

func (s *Stenographer) printBanner(text string, bannerCharacter string) {
	fmt.Println(text)
	fmt.Println(strings.Repeat(bannerCharacter, len(text)))
}

func (s *Stenographer) printNewLine() {
	fmt.Println("")
}

func (s *Stenographer) printDelimiter() {
	fmt.Println(s.colorize(grayColor, "%s", strings.Repeat("-", 30)))
}

func (s *Stenographer) print(indentation int, format string, args ...interface{}) {
	fmt.Print(s.indent(indentation, format, args...))
}

func (s *Stenographer) println(indentation int, format string, args ...interface{}) {
	fmt.Println(s.indent(indentation, format, args...))
}

func (s *Stenographer) indent(indentation int, format string, args ...interface{}) string {
	var text string

	if len(args) > 0 {
		text = fmt.Sprintf(format, args...)
	} else {
		text = format
	}

	stringArray := strings.Split(text, "\n")
	padding := ""
	if indentation >= 0 {
		padding = strings.Repeat("  ", indentation)
	}
	for i, s := range stringArray {
		stringArray[i] = fmt.Sprintf("%s%s", padding, s)
	}

	return strings.Join(stringArray, "\n")
}
