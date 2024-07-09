package utils

import "bufio"

func SplitBySeparator(separator []byte) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	isSep := func(data []byte) bool {
		if len(data) != len(separator) {
			return false
		}
		for i := range len(separator) {
			if data[i] != separator[i] {
				return false
			}
		}
		return true
	}

	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			return 0, nil, bufio.ErrFinalToken
		}
		for i := range len(data) {
			upToIdx := i + len(separator)
			if upToIdx > len(data) {
				break
			}
			if isSep(data[i:upToIdx]) {
				return i + len(separator), data[:i], nil
			}
		}
		return len(data), data, bufio.ErrFinalToken
	}
	return split
}
