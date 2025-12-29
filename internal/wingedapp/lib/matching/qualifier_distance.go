package matching

import (
	"context"
	"errors"
	"fmt"

	"github.com/umahmood/haversine"
)

// newDistanceQualifier creates a new distance qualifier,
// and attaches it to the QualifierResults.
func newDistanceQualifier(qr *QualifierResults) *Qualifier {
	if qr.Distance != nil {
		panic(distanceQualifier + " already set")
	}
	qr.Distance = newQualifier(distanceQualifier)
	return qr.Distance
}

// distanceQualifier checks if the distance between two users is acceptable.
func (l *Logic) distanceQualifier(ctx context.Context, params *QualifierParameters) *Qualifier {
	q := newDistanceQualifier(params.QualifierResults)

	userA, userB := params.Users()

	// guard: both users must have valid location
	if !userA.Latitude.Valid || !userA.Longitude.Valid ||
		!userB.Latitude.Valid || !userB.Longitude.Valid {
		return q.SetError(ErrLocationNotSet)
	}

	pointA := haversine.Coord{Lat: userA.Latitude.Float64, Lon: userA.Longitude.Float64}
	pointB := haversine.Coord{Lat: userB.Latitude.Float64, Lon: userB.Longitude.Float64}
	_, distKM := haversine.Distance(pointA, pointB)

	q.Telemetry["distance_km"] = distKM

	cfg := params.config

	// base case: within radius
	if distKM <= cfg.LocationRadiusKM {
		return q
	}

	// attempt adaptive expansion if enabled
	for _, expandedRadius := range cfg.LocationAdaptiveExpansion {
		q.Telemetry["attempt_adaptive_expansion"] = expandedRadius
		if int64(distKM) <= expandedRadius {
			q.Telemetry["adaptive_expansion_passed"] = expandedRadius
			return q
		}
	}

	// fail - distance exceeds all allowed radii
	return q.SetError(errors.Join(
		ErrDistanceExceeds,
		fmt.Errorf("distance %.2f km exceeds max allowed radius %v km", distKM, cfg.LocationRadiusKM),
		fmt.Errorf("adaptive radius considered: %v", cfg.LocationAdaptiveExpansion),
	))
}
