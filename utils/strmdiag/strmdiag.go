package strmdiag

import (
	"fmt"
)

func humanize(count float64) string {
	units := []string{"", "K", "M", "G", "T"}
	cur := 0
	for ; count > 1000 && cur < len(units); cur++ {
		count /= 1000
	}
	return fmt.Sprintf("%.02f%s", count, units[cur])
}
