package factory

import "github.com/google/uuid"

func RandomName() string {
	return uuid.NewString()
}

func RandomEmail() string {
	return uuid.NewString() + "@example.com"
}

func TestEmail() string {
	return "test@gmail.com"
}
