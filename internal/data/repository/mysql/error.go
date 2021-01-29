package mysql

import "github.com/shipa988/hw_otus_architect/internal/domain/entity"

var _ entity.ValidateError =(*MySqlError)(nil)

type MySqlError struct {
	Msg string
	isLoginExist,wrongLenUserorPas bool
}

func (e MySqlError) Error() string {
	return e.Msg
}

func (e *MySqlError) IsLoginExist() bool  {
	return e.isLoginExist
}
func (e *MySqlError) SetLoginExist()  {
	e.isLoginExist=true
}

func (e *MySqlError) IsWrongLenUserorPas() bool  {
	return e.wrongLenUserorPas
}
func (e *MySqlError) SetWrongLenUserorPas()  {
	e.wrongLenUserorPas=true
}