// Package windowscmd contains Windows command-line parsing helpers.
package windowscmd

import "strings"

func SplitArgs(value string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	started := false
	for index := 0; index < len(value); {
		if (value[index] == ' ' || value[index] == '\t') && !inQuotes {
			if started {
				args = append(args, current.String())
				current.Reset()
				started = false
			}
			index++
			continue
		}

		backslashes := 0
		for index < len(value) && value[index] == '\\' {
			backslashes++
			index++
		}
		if index < len(value) && value[index] == '"' {
			current.WriteString(strings.Repeat("\\", backslashes/2))
			if backslashes%2 == 0 {
				inQuotes = !inQuotes
			} else {
				current.WriteByte('"')
			}
			started = true
			index++
			continue
		}
		if backslashes > 0 {
			current.WriteString(strings.Repeat("\\", backslashes))
			started = true
			continue
		}
		current.WriteByte(value[index])
		started = true
		index++
	}
	if started {
		args = append(args, current.String())
	}
	return args
}
