package strmtest

import (
	"fmt"
	"regexp"
)

func MatchError(match string, err error) bool {
	if match == "" {
		return err == nil
	}

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	re := regexp.MustCompile(match)
	return re.MatchString(errStr)
}

func MatchPanic(match string, p interface{}) bool {
	if match == "" {
		return p == nil
	}
	str := ""
	if p != nil {
		str = fmt.Sprint(p)
	}
	re := regexp.MustCompile(match)
	return re.MatchString(str)
}
