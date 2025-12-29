package db

type TransactorSupabaseAuth struct {
	*Transactor
}

// NewTransactorSupabaseAuth creates a new TransactorSupabaseAuth instance,
// embedding the provided Transactor.
// This is just so we can differentiate our struct types in DI containers.
func NewTransactorSupabaseAuth(t *Transactor) *TransactorSupabaseAuth {
	if t == nil {
		panic("TransactorSupabaseAuth: nil Transactor provided")
	}
	return &TransactorSupabaseAuth{Transactor: t}
}
