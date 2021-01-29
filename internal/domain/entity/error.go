package entity

const (
	ErrWrongUserOrPass = "Wrong user or password"
	ErrLenUserorPas = "Wrong user or password length"
	ErrLoginExist = "User with same login exist"
)


type ValidateError interface {
	error
	IsLoginExist() bool
	IsWrongLenUserorPas() bool
}
