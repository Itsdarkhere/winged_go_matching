package factory

import (
	"testing"

	"github.com/aarondl/sqlboiler/v4/boil"
)

/*
	This package provides a generic factory pattern for initializing,
	and setting defaults for types in tests.
*/

// Instantiator is an interface that defines methods for initializing and setting defaults for a type T.
type Instantiator[T any] interface {
	New(t *testing.T, db boil.ContextExecutor) T // New initializes a new instance of type T for testing.
	SetRequiredFields() T                        // SetRequiredFields sets the required default fields for the instance.
	Save() T                                     // Save persists the instance to the database or any storage.
	IsValid() bool                               // IsValid checks if the instance is valid for use in tests.
	EnsureFKDeps()                               // EnsureFKDeps ensures factory FK deps are persisted
}

// Entity is a generic struct that holds a type T which implements the Instantiator interface.
type Entity[T Instantiator[T]] struct {
	T T
}

func NewEntity[T Instantiator[T]](t T) *Entity[T] {
	return &Entity[T]{T: t}
}

// New initializes the entity with a test instance and returns it.
func (e *Entity[T]) New(t *testing.T, db boil.ContextExecutor) T {
	entity := e.T.New(t, db)
	entity.EnsureFKDeps()
	entity = entity.SetRequiredFields()
	return entity.Save()
}
