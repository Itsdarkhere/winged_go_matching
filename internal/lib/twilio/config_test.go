package twilio

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRandomDigitCode test it can generate a random digit code of specified length,
// and the generated string is numeric.
func TestRandomDigitCode(t *testing.T) {
	cases := []int{1, 4, 6, 8}
	for _, n := range cases {
		t.Run(fmt.Sprintf("length_%d", n), func(t *testing.T) {
			t.Parallel()
			got := RandomDigitCode(n)
			require.Len(t, got, n)
			require.Regexp(t, "^[0-9]+$", got)

			val, err := strconv.Atoi(got)
			require.NoError(t, err)
			require.GreaterOrEqual(t, val, 0)
			require.Less(t, val, int(math.Pow10(n)))
		})
	}
}
