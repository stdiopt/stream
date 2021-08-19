package strmtest

import (
	"fmt"
	"regexp"
)

func matchError(match string, err error) bool {
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

func matchPanic(match string, p interface{}) bool {
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

type testSender struct {
	sendFunc  func(interface{}) error
	closeFunc func() error
}

func (s testSender) Send(v interface{}) error {
	if s.sendFunc != nil {
		return s.sendFunc(v)
	}
	return nil
}

func (s testSender) Close() error {
	if s.closeFunc != nil {
		return s.closeFunc()
	}
	return nil
}