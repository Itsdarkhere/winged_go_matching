package db

type TransactorAIBackend struct {
	*Transactor
}

// NewTransactorAIBackend creates a new TransactorAIBackend instance,
// embedding the provided Transactor.
// This is just so we can differentiate our struct types in DI containers.
func NewTransactorAIBackend(t *Transactor) *TransactorAIBackend {
	if t == nil {
		panic("TransactorBackendApp: nil Transactor provided")
	}
	return &TransactorAIBackend{Transactor: t}
}
