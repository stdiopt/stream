package stream

import (
	"regexp"
	"testing"
)

func Test_procName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "returns a procname in the format",
			want: `\w+\.\w+\(.*:\d+\)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.want)
			if got := procName(); !re.MatchString(got.String()) {
				t.Errorf("procName() = %v, want %v", got, tt.want)
			}
		})
	}
}
