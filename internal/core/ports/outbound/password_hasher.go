package outbound

//go:generate go run github.com/vektra/mockery/v2 --name=PasswordHasher

// PasswordHasher abstracts password hashing and verification.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hashed, plain string) error // returns domain.ErrInvalidCredentials on mismatch
}
