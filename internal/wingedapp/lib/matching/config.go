package matching

import (
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// UpdateConfig updates the match configuration with the provided fields.
func (l *Logic) UpdateConfig(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *UpdateMatchConfig,
) (*Config, error) {
	// Validate the update request
	if err := updater.Validate(); err != nil {
		return nil, err
	}

	config, err := l.configStorer.Update(ctx, exec, updater)
	if err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}

	return config, nil
}
