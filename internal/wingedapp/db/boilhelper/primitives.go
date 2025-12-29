package boilhelper

import "github.com/aarondl/sqlboiler/v4/types"

func Int64(d *types.Decimal) int64 {
	if d == nil {
		return 0
	}
	i, _ := d.Int64()
	return i
}
