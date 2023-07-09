package main

import (
	"database/sql"

	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DataWorker struct {
	tableName string
	db        *sql.DB
}

type dataRow struct {
	Value       float32
	Description string
}

// TeleBotGo/finances.db
func NewDataWorker() DataWorker {
	database, err := sql.Open("sqlite3", "/home/anton/TeleBotGo/finances.db")
	if err != nil {
		log.Fatal(err)
	}
	return DataWorker{
		tableName: "Finance",
		db:        database,
	}
}

func (dw *DataWorker) CreateTableIfNotExists() {
	statement, err := dw.db.Prepare("CREATE TABLE IF NOT EXISTS " + dw.tableName + "(id INTEGER PRIMARY KEY AUTOINCREMENT, idUser TEXT, value REAL, description TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}

}

func (dw *DataWorker) CreateTable() {
	statement, err := dw.db.Prepare("CREATE TABLE " + dw.tableName + "(id INTEGER PRIMARY KEY AUTOINCREMENT, idUser TEXT, value REAL, description TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func (dw *DataWorker) GetData(userId int) ([]dataRow, error) {
	rows, err := dw.db.Query("SELECT value, description FROM "+dw.tableName+" WHERE idUser = ?", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []dataRow
	for rows.Next() {
		var row dataRow
		err = rows.Scan(&row.Value, &row.Description)
		if err != nil {
			return nil, err
		}
		data = append(data, row)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return data, nil
}

func (dw *DataWorker) InsertData(userId int, value float32, description string) error {
	statement, err := dw.db.Prepare("INSERT INTO " + dw.tableName + " (idUser, value, description) VALUES (?,?,?)")
	if err != nil {
		log.Fatal(err)
		return err
	}
	_, err = statement.Exec(userId, value, description)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (dw *DataWorker) DelData(userId int) (dataRow, error) {
	rows, err := dw.db.Query("SELECT id, value, description  FROM " + dw.tableName + " WHERE id = (SELECT MAX(id) FROM "+ dw.tableName +" WHERE idUser = ?)", userId)
	if err != nil {
		log.Println(">")
		return dataRow{}, err
	}
	defer rows.Close()
	var id int
	var delRow dataRow
	for rows.Next() {
		if err := rows.Scan(&id, &delRow.Value, &delRow.Description); err != nil {
			log.Println(err)
			return dataRow{}, err
		}
	}


	statement, err := dw.db.Prepare("DELETE FROM " + dw.tableName + " WHERE id = ?")
	if err != nil {
		return dataRow{}, err
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err!= nil {
		return dataRow{}, err
	}
    

    return delRow, nil
}
