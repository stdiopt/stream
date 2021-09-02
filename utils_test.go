package stream

import (
	"context"
	"fmt"
	"regexp"
	"time"
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

type fakeContext struct {
	s string
	context.Context
}

type fakeConsumer struct {
	s string
	consumer
}

type fakeSender struct {
	s string
	sender
}

type contextHelper struct {
	context.Context
	cancel func()
}

func helperCanceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	return contextHelper{
		ctx,
		cancel,
	}
}

func helperTimeoutContext(d time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.TODO(), d)
	return contextHelper{
		ctx,
		cancel,
	}
}
