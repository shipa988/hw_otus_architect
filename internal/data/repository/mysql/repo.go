package mysql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/config"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"time"
)

const (
	ErrAdd = "can't add new rows to table"
	ErrGet = "can't get profiles from db"
)

const (
	connectErr    = `can't connect to mysql'`
	disconnectErr = `can't disconnect from mysql'`
	insertErr     = `can't insert to mysql'`
)

var _ entity.UserRepository = (*mySqlRepo)(nil)
var _ entity.ProfileRepository = (*mySqlRepo)(nil)
var _ entity.UserAuthRepository = (*mySqlRepo)(nil)

type mySqlRepo struct {
	master *sql.DB
	slaves []*sql.DB
}

func NewMySqlRepo() *mySqlRepo {
	return &mySqlRepo{}
}
func (repo *mySqlRepo) getSlave(userId uint64) *sql.DB {
	return repo.slaves[userId%uint64(len(repo.slaves))]
}
func (repo *mySqlRepo) connect(ctx context.Context,provider, login, password, addr, name string) (db *sql.DB,err error) {
	db, err = sql.Open(provider, fmt.Sprintf("%v:%v@tcp(%v)/%v", login, password, addr, name))
	if err != nil {
		return nil,errors.Wrap(err, connectErr)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)

	db.Stats()
	err = db.PingContext(ctx)
	if err != nil {
		return nil,errors.Wrap(err, connectErr)
	}
	log.Info("connected to mysql server [%v]", addr)
	return db,nil
}
func (repo *mySqlRepo) Connect(ctx context.Context, cfg config.DB) (err error) {
	repo.master,err=repo.connect(ctx,cfg.Provider, cfg.Login, cfg.Password, cfg.Master, cfg.Name)
	if err != nil {
		return err
	}
	for _, slave := range cfg.Slaves {
		slavedb,err:=repo.connect(ctx,cfg.Provider, cfg.Login, cfg.Password, slave, cfg.Name)
		if err != nil {
			return err
		}
		repo.slaves = append(repo.slaves, slavedb)
	}
	return nil
}

func (repo *mySqlRepo) IsSubscribed(ctx context.Context, userId uint64, subscibeId uint64) (bool, error) {
	cnt := 0
	err := repo.master.QueryRowContext(ctx, `SELECT count(*) FROM Friends where UserId=? and FriendId=?;`, userId, subscibeId).Scan(&cnt)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, errors.Wrap(err, "get user by id error")
	case cnt == 0:
		return false, nil
	default:
		return true, nil
	}
}

func (repo *mySqlRepo) Subscribe(ctx context.Context, fromId uint64, toId uint64) error {
	_, err := repo.master.ExecContext(ctx, "INSERT INTO Friends (`UserId`,`FriendId`)VALUES(?,?);", fromId, toId)
	if err != nil {
		return errors.Wrap(err, "subscribe error")
	}
	return nil
}

func (repo *mySqlRepo) UnSubscribe(ctx context.Context, fromId uint64, toId uint64) error {
	_, err := repo.master.ExecContext(ctx, "DELETE from Friends where `UserId`=? and`FriendId`=?;", fromId, toId)
	if err != nil {
		return errors.Wrap(err, "unsubscribe error")
	}
	return nil
}

func (repo *mySqlRepo) GetFriendsById(ctx context.Context, id uint64, limit int, lastID uint64) ([]entity.User, error) {
	rows, err := repo.getSlave(id).QueryContext(ctx, `select allp.Id,allp.Name,allp.SurName,allp.Age,allp.Gen,allp.Interest,allp.City  from (select FriendId from Friends where UserId=?) fr left join (SELECT Id,Name,SurName,Age,Gen,Interest,City FROM Profiles) allp on fr.FriendId = allp.Id where allp.Id>? order by allp.Id limit ?;`, id, lastID, limit)
	if err != nil && err != sql.ErrNoRows {
		return nil, SQLError(err, ErrGet)
	}
	defer rows.Close()
	return repo.rowsToUsers(rows, ErrGet)
}

func (repo *mySqlRepo) FilterByNameSurName(ctx context.Context, myuId uint64, name, surname string, limit int, lastID uint64) ([]entity.User, error) {
	rows, err := repo.getSlave(myuId).QueryContext(ctx, `SELECT Id,Name,SurName,Age,Gen,Interest,City FROM Profiles where Id<>? and Name like concat(?, '%') and SurName like concat(?, '%') and Id>? order by Id limit ?;`, myuId, name, surname, lastID, limit)
	if err != nil && err != sql.ErrNoRows {
		return nil, SQLError(err, ErrGet)
	}
	defer rows.Close()
	return repo.rowsToUsers(rows, ErrGet)
}

func (repo *mySqlRepo) GetUserById(ctx context.Context, id uint64) (entity.User, error) {
	user := entity.User{}
	err := repo.getSlave(id).QueryRowContext(ctx, `SELECT Id,Name,SurName,Age,Gen,Interest,City FROM Profiles where Id=?;`, id).Scan(&user.Id, &user.Name, &user.SurName, &user.Age, &user.Gen, &user.Interest, &user.City)

	switch {
	case err == sql.ErrNoRows:
		return user, nil
	case err != nil:
		return user, errors.Wrap(err, "get user by id error")
	default:
		return user, nil
	}
}

func (repo *mySqlRepo) LogOff(ctx context.Context, id uint64, uuid string) (err error) {
	_, err = repo.master.ExecContext(ctx, "DELETE from Seanses where `UserId`=?", id)
	if err != nil {
		return errors.Wrap(err, "logoff error")
	}
	return nil
}

func (repo *mySqlRepo) SignIn(ctx context.Context, uuid string, id uint64) error {
	_, err := repo.master.ExecContext(ctx, "INSERT INTO Seanses (`UserId`,`Uuid`)VALUES(?,?);", id, uuid)
	if err != nil {
		return errors.Wrap(err, "signin error")
	}
	return nil
}

func (repo *mySqlRepo) IsSignIn(ctx context.Context, uuid string) (id uint64, ok bool, err error) {
	err = repo.master.QueryRowContext(ctx, `SELECT UserId FROM Seanses where Uuid=?;`, uuid).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return 0, false, nil
	case err != nil:
		return 0, false, errors.Wrap(err, "get user by id error")
	default:
		return id, true, nil
	}
}

func (repo *mySqlRepo) Register(ctx context.Context, login, name, hash string) (uint64, error) {
	res, err := repo.master.ExecContext(ctx, "INSERT INTO Users (Login,PassHash) VALUES(?,?)", login, hash)
	if err != nil {
		return 0, errors.Wrap(err, "register error")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "register error")
	}
	user := entity.User{
		Id:   uint64(id),
		Name: name,
	}
	err = repo.SaveUser(ctx, user)
	if err != nil {
		return 0, errors.Wrap(err, "register error")
	}
	return uint64(id), nil
}

func (repo *mySqlRepo) GetUserAuth(ctx context.Context, login string) (uint64, string, error) {
	var id uint64
	var hash string
	err := repo.master.QueryRowContext(ctx, `SELECT Id,PassHash FROM Users where Login=?;`, login).Scan(&id, &hash)

	if err != nil {
		return 0, "", errors.Wrap(err, "get user auth by login error")
	}
	return id, hash, nil
}

func (repo *mySqlRepo) SaveUser(ctx context.Context, user entity.User) error {
	_, err := repo.master.ExecContext(ctx, "INSERT INTO Profiles (`Id`, `Name`, `SurName`, `Age`, `Gen`, `Interest`, `City`) VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE `Id`=?,`Name`=?,`SurName`=?,`Age`=?,`Gen`=?,`Interest`=?,`City`=?", user.Id, user.Name, user.SurName, user.Age, user.Gen, user.Interest, user.City, user.Id, user.Name, user.SurName, user.Age, user.Gen, user.Interest, user.City)
	if err != nil {
		return errors.Wrap(err, "save User error")
	}
	return nil
}

func (repo *mySqlRepo) Validate(ctx context.Context, login, pass string) (bool, error) {
	if len(login) == 0 || len(login) > 20 {
		err := &MySqlError{
			Msg: "Length of login is invalid",
		}
		err.SetWrongLenUserorPas()
		return false, err
	}
	if len(pass) == 0 || len(pass) > 20 {
		err := &MySqlError{
			Msg: "Length of pass is invalid",
		}
		err.SetWrongLenUserorPas()
		fmt.Println()
		return false, err
	}
	var id uint64
	err := repo.master.QueryRowContext(ctx, `select Id from Users where Login=?;`, login).Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return true, nil
	case err != nil:
		return false, errors.Wrap(err, "validate error")
	default:
		err := &MySqlError{
			Msg: "login already exist",
		}
		err.SetLoginExist()
		return false, err
	}
}

func (repo *mySqlRepo) SaveUsersBatch(ctx context.Context, users []entity.User) error {
	/*	sqlStr := "insert into public.users (id, name, rate) values "
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
		}*/
	return nil
}

func (repo *mySqlRepo) rowsToUsers(rows *sql.Rows, errorString string) ([]entity.User, error) {
	var users []entity.User
	for rows.Next() {
		user := entity.User{}
		err := rows.Scan(&user.Id, &user.Name, &user.SurName, &user.Age, &user.Gen, &user.Interest, &user.City)
		switch {
		case err == sql.ErrNoRows:
			return nil, nil
		case err != nil:
			return nil, errors.Wrap(err, "convert rows to users error")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, SQLError(err, errorString)
	}
	return users, nil
}

func SQLError(err error, message string) error {
	switch err {
	case sql.ErrNoRows:
		return nil
	default:
		return errors.Wrap(err, message)
	}
}
