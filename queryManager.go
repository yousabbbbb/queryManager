package queryManager

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

type User struct {
	Id          int
	Name        string
	Username    string
	Surname     string
	Description string
}

var (
	host     = ""
	port     = 5432
	username = ""
	password = ""
	dbname   = ""
)

func openConnection() (*sql.DB, error) {
	log.SetFlags(0)
	connectString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, username, password, dbname)
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		log.Printf("couldn't open the database %v\n", err)
		return nil, err
	}
	defer db.Close()
	return db, nil
}

func exists(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		log.Printf("error while opening the database %v\n", err)
		return -1
	}
	defer db.Close()

	statement := fmt.Sprintf(`SELECT "id" FROM "users" WHERE username = %s`, username)
	rows, err := db.Query(statement)
	if err != nil {
		log.Printf("error while searching for uesr %v\n", err)
		return -1
	}

	userId := -1

	for rows.Next() {
		var Id int
		err := rows.Scan(&Id)
		if err != nil {
			log.Printf("error while searching for uesr %v\n", err)
			return userId
		}
	}

	defer rows.Close()
	return userId
}

func AddUser(d User) int {
	d.Username = strings.ToLower(d.Username)

	db, err := openConnection()
	if err != nil {
		log.Printf("error while opening the database %v\n", err)
		return -1
	}
	defer db.Close()

	statement := fmt.Sprintf(`INSERT INTO "users" ("username") VALUES ($1)`)

	_, err = db.Exec(statement, d.Username)
	if err != nil {
		log.Printf("error while insering user to database %v\n", err)
		return -1
	}
	userId := exists(username)
	if userId == -1 {
		return userId
	}

	statement = fmt.Sprintf(`INSERT INTO "userdata" (userid, name, surname, description) VALUES ($1, $2, $3, $4)`)
	_, err = db.Exec(statement, userId, d.Name, d.Surname, d.Description)
	if err != nil {
		log.Println("error while inserting userdata to the database", err)
		return -1
	}
	return userId
}

func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	statement := fmt.Sprintf(`SELECT "username" FROM "users" WHERE id = %d`, id)
	rows, err := db.Query(statement)
	if err != nil {
		return err
	}
	defer rows.Close()
	var username string
	for rows.Next() {
		err := rows.Scan(&username)
		if err != nil {
			return err
		}
	}

	userId := exists(username)
	if userId != id {
		log.Fatalf("user with id %d doesn't exist", id)
	}

	deleteStatement := fmt.Sprintf(`delete from "userdata" where userid=$1`)
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	deleteStatement = fmt.Sprintf(`delete from "users" WHERE id=$1`)
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}
	return nil
}

func ListUsers() ([]User, error) {
	userData := []User{}
	db, err := openConnection()
	if err != nil {
		return userData, err
	}
	defer db.Close()

	statement := fmt.Sprint(`SELECT "id", "name","username","surname","description" FROM "users", "userdata" where users.id = userdata.userid`)
	rows, err := db.Query(statement)
	if err != nil {
		return userData, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          int
			name        string
			username    string
			surname     string
			description string
		)
		err := rows.Scan(&id, &name, &username, &surname, &description)
		if err != nil {
			return userData, err
		}
		temp := User{Id: id, Name: name, Username: username, Surname: surname, Description: description}
		userData = append(userData, temp)
	}
	return userData, nil
}

func UpdateUser(d User) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	userId := exists(d.Username)
	if userId == -1 {
		return errors.New("User does not exist")
	}
	d.Id = userId

	statement := `update "userdata" set "name"=$1, "surname"=$2, "description"=$3 where "userid"=$4`
	_, err = db.Exec(statement, d.Name, d.Surname, d.Description, d.Id)
	if err != nil {
		return err
	}

	return nil
}
