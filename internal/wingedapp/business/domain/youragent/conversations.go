package youragent

import (
	"context"
	"fmt"
)

// PromptConversations fetches past conversations for a user.
func (b *Business) PromptConversations(ctx context.Context, f *UserAIConvoQueryFilter) (*UserAIConvos, error) {
	u, err := b.storer.UserAIConvos(ctx, b.trans.DB(), f)
	if err != nil {
		return nil, fmt.Errorf("UserAIConvos: %w", err)
	}

	return &UserAIConvos{
		Results: u,
		Paging:  f.Pagination,
	}, nil
}
