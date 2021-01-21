package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/config"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"time"
)

const (
	ErrAdd = "can't add new rows to table"
	ErrGet = "can't get currencies from db"
)

const (
	connectErr    = `can't connect to clickhouse'`
	disconnectErr = `can't disconnect from clickhouse'`
	insertErr     = `can't insert to clickhouse'`
)

var _ entity.UserRepository = (*MySqlRepo)(nil)

type MySqlRepo struct {
	db *sql.DB
}

func (repo *MySqlRepo) FilterByNameSurName(ctx context.Context, name, surname string, limit int, lastID uint) ([]entity.User, error) {
	panic("implement me")
}

func NewMySqlRepo() *MySqlRepo {
	return &MySqlRepo{}
}

func (repo *MySqlRepo) Connect(ctx context.Context, cfg config.DB) (err error) {
	repo.db, err = sql.Open(cfg.Provider, fmt.Sprintf("%v:%v@/%v", cfg.Login, cfg.Password, cfg.Name))
	if err != nil {
		return errors.Wrap(err, connectErr)
	}
	// See "Important settings" section.
	repo.db.SetConnMaxLifetime(time.Minute * 3)
	repo.db.SetMaxOpenConns(100)
	repo.db.SetMaxIdleConns(100)

	repo.db.Stats()
	err = repo.db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, connectErr)
	}
	log.Info("connected to mysql server [%v:%v]", cfg.Address, cfg.Port)
	return nil
}

func (repo *MySqlRepo) GetUserById(ctx context.Context, id uint) (entity.User, error) {
	panic("implement me")
}

func (repo *MySqlRepo) SaveUsersBatch(ctx context.Context, users []entity.User) (error) {
	sqlStr := "insert into public.users (id, name, rate) values "
	var vals []interface{}
	for i, row := range cs {
		sqlStr += fmt.Sprintf("($%v,$%v,$%v),", (i*3)+1, (i*3)+2, (i*3)+3)
		vals = append(vals, row.ID, row.Name, row.Value/float64(row.Nominal))
	}
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	sqlStr += " on conflict (id) do UPDATE SET (rate,insert_dt)=(EXCLUDED.rate,now());"
	stmt, err := repo.db.Prepare(sqlStr)
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	result, err := stmt.ExecContext(ctx, vals...)
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	if rows != int64(len(cs)) {
		return errors.Wrapf(err, ErrAdd)
	}
	return nil
}

//Проверить входящие данные пользователя ...
func (account *Account) Validate() (map[string] interface{}, bool) {

	if !strings.Contains(account.Email, "@") {
		return u.Message(false, "Email address is required"), false
	}

	if len(account.Password) < 6 {
		return u.Message(false, "Password is required"), false
	}

	//Email должен быть уникальным
	temp := &Account{}

	//проверка на наличие ошибок и дубликатов электронных писем
	err := GetDB().Table("accounts").Where("email = ?", account.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error. Please retry"), false
	}
	if temp.Email != "" {
		return u.Message(false, "Email address already in use by another user."), false
	}

	return u.Message(false, "Requirement passed"), true
}


func (repo *MySqlRepo) SaveUser(ctx context.Context, user entity.User) (uint, error) {
	sqlStr := "insert into public.users (id, name, rate) values "
	var vals []interface{}
	for i, row := range cs {
		sqlStr += fmt.Sprintf("($%v,$%v,$%v),", (i*3)+1, (i*3)+2, (i*3)+3)
		vals = append(vals, row.ID, row.Name, row.Value/float64(row.Nominal))
	}
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	sqlStr += " on conflict (id) do UPDATE SET (rate,insert_dt)=(EXCLUDED.rate,now());"
	stmt, err := repo.db.Prepare(sqlStr)
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	result, err := stmt.ExecContext(ctx, vals...)
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	if rows != int64(len(cs)) {
		return errors.Wrapf(err, ErrAdd)
	}
	return nil
}

func (repo *MySqlRepo) SetAll(ctx context.Context, cs []*entity.Currency) error {
	sqlStr := "insert into public.currency (id, name, rate) values "
	var vals []interface{}
	for i, row := range cs {
		sqlStr += fmt.Sprintf("($%v,$%v,$%v),", (i*3)+1, (i*3)+2, (i*3)+3)
		vals = append(vals, row.ID, row.Name, row.Value/float64(row.Nominal))
	}
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	sqlStr += " on conflict (id) do UPDATE SET (rate,insert_dt)=(EXCLUDED.rate,now());"
	stmt, err := repo.db.Prepare(sqlStr)
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	result, err := stmt.ExecContext(ctx, vals...)
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, ErrAdd)
	}

	if rows != int64(len(cs)) {
		return errors.Wrapf(err, ErrAdd)
	}
	return nil
}

func (repo *PGSRepo) GetByID(ctx context.Context, id string) (*entity.Currency, error) {
	row := repo.db.QueryRowContext(ctx, `select id, name, rate 
												from public.currency where id=$1;`, id)
	if row == nil {
		return nil, nil
	}

	c := entity.Currency{Nominal: 1}
	err := row.Scan(&c.ID, &c.Name, &c.Value)
	if err != nil {
		return nil, SQLError(err, ErrGet)
	}

	return &c, nil
}

func (repo *PGSRepo) GetPage(ctx context.Context, limit, offset int) ([]*entity.Currency, error) {
	rows, err := repo.db.QueryContext(ctx, `select id, name, rate 
												from public.currency order by id limit $1 offset $2;`, limit, offset)
	if err != nil && err != sql.ErrNoRows {
		return nil, SQLError(err, ErrGet)
	}
	defer rows.Close()
	return repo.rowsToCurrencies(rows, ErrGet)
}

func (repo *PGSRepo) GetLazy(ctx context.Context, limit int, lastID string) ([]*entity.Currency, error) {
	rows, err := repo.db.QueryContext(ctx, `select id, name, rate
												from public.currency where id>$1 order by id limit $2;`, lastID, limit)
	if err != nil && err != sql.ErrNoRows {
		return nil, SQLError(err, ErrGet)
	}
	defer rows.Close()
	return repo.rowsToCurrencies(rows, ErrGet)
}

func (repo *PGSRepo) Connect(ctx context.Context, dsn string) (err error) {
	err = repo.db.PingContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to db: %v", dsn)
	}
	return nil
}

func (repo *PGSRepo) Close() error {
	return repo.db.Close()
}

func (repo *PGSRepo) rowsToCurrencies(rows *sql.Rows, errorString string) ([]*entity.Currency, error) {
	var currencies []*entity.Currency
	for rows.Next() {
		c := entity.Currency{Nominal: 1}
		err := rows.Scan(&c.ID, &c.Name, &c.Value)
		if err != nil {
			return nil, SQLError(err, errorString)
		}
		currencies = append(currencies, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, SQLError(err, errorString)
	}
	return currencies, nil
}

func SQLError(err error, message string) error {
	switch err {
	case sql.ErrNoRows:
		return nil
	default:
		return errors.Wrap(err, message)
	}
}
