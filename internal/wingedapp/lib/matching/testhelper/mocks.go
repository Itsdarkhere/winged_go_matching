package testhelper

import (
	"context"
	"encoding/json"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/matchingfakes"
)

// MockSuccessQualitativeMatch returns a mock ProfileGetter that always returns a successful profile.
func MockSuccessQualitativeMatch(t *testing.T) *matchingfakes.FakeQualitativeQuantifier {
	f := &matchingfakes.FakeQualitativeQuantifier{}
	var m matching.MatchCompatibilityResult
	if err := json.Unmarshal([]byte(mockSuccessPopulationResponse), &m); err != nil {
		t.Fatalf("failed to unmarshal mock profile: %v", err)
	}
	f.QualifyReturns(&m, nil)
	return f
}

// MockPublicURLer returns a mock publicURLer that converts storage keys to public URLs.
// By default it prepends "https://public.supabase.io/" to the key.
func MockPublicURLer() *matchingfakes.FakePublicURLer {
	f := &matchingfakes.FakePublicURLer{}
	f.PublicURLStub = func(_ context.Context, key string) (string, error) {
		return "https://public.supabase.io/" + key, nil
	}
	return f
}
