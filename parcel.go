package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	lastParcelNum := 0
	res, err := s.db.Exec("INSERT INTO parcel (status, client, address, created_at) VALUES (:status, :client, :address, :created_at)",
		sql.Named("status", ParcelStatusRegistered),
		sql.Named("client", p.Client),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return lastParcelNum, err
	}
	lastNum, _ := res.LastInsertId()

	// верните идентификатор последней добавленной записи
	return int(lastNum), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка

	row := s.db.QueryRow("SELECT client, status, address, created_at FROM parcel WHERE number = :number", sql.Named("number", number))

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	p.Number = number
	err := row.Scan(&p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	var res []Parcel
	rows, err := s.db.Query("SELECT number, status, address, created_at, client FROM parcel WHERE client = :client", sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}
		p := Parcel{}
		err := rows.Scan(&p.Number, &p.Status, &p.Address, &p.CreatedAt, &p.Client)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	// заполните срез Parcel данными из таблицы

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("error setting status for parcel %d - %w", number, err)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	res, err := s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number AND status = :status",
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return fmt.Errorf("error setting address for parcel %d - %w", number, err)
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("error setting address for parcel %d", number)
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered

	res, err := s.db.Exec("DELETE FROM parcel WHERE number = :number AND status = :status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return fmt.Errorf("error deleting parcel %d - %w", number, err)
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("error deleting parcel %d", number)
	}
	return nil
}
