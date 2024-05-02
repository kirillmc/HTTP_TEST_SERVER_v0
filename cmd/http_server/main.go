package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	idColumn = "id"
	name     = "name"
	surname  = "surname"
	email    = "email"
	avatar   = "avatar"
	login    = "login"
	password = "password"
	role     = "role"
	weight   = "weight"
	height   = "height"
	locked   = "locked"
)

// User структура представляет пользователя
type User struct {
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Email    string  `json:"email"`
	Avatar   string  `json:"avatar"`
	Login    string  `json:"login"`
	Password string  `json:"password"`
	Role     int32   `json:"role"`
	Weight   float64 `json:"weight"`
	Height   float64 `json:"height"`
	Locked   bool    `json:"locked"`
}

type UserToGet struct {
	Id       int64   `json:"id"`
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Email    string  `json:"email"`
	Avatar   string  `json:"avatar"`
	Login    string  `json:"login"`
	Password string  `json:"password"`
	Role     int32   `json:"role"`
	Weight   float64 `json:"weight"`
	Height   float64 `json:"height"`
	Locked   bool    `json:"locked"`
}

type CreateResponse struct {
	Id int64 `json:"id"`
}

var db *sql.DB

func main() {
	// Загрузка значений из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Подключение к базе данных PostgreSQL
	db, err = sql.Open("postgres", os.Getenv("PG_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUser)
	log.Println("Server is serving on: localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createUser(w, r)
	case http.MethodGet:
		getUsers(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	query := fmt.Sprintf("SELECT %s ,%s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM users", idColumn, name, surname, email, avatar, login, password, role, weight, height, locked)
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []UserToGet{}
	for rows.Next() {
		var user UserToGet
		err := rows.Scan(&user.Id, &user.Name, &user.Surname, &user.Email, &user.Avatar, &user.Login, &user.Password,
			&user.Role, &user.Weight, &user.Height, &user.Locked)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUserById(w, r)
	case http.MethodPut:
		updateUserById(w, r)
	case http.MethodDelete:
		deleteUserByID(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {

	// Декодируем JSON из тела запроса в структуру User
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var id int64
	// Вставляем данные пользователя в базу данных
	query := fmt.Sprintf("INSERT INTO users (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5,$6, $7, $8, $9, $10 ) RETURNING id ", name, surname, email, avatar, login, password, role, weight, height, locked)

	err = db.QueryRow(query,
		user.Name,
		user.Surname,
		user.Email,
		user.Avatar,
		user.Login,
		user.Password,
		user.Role,
		user.Weight,
		user.Height,
		user.Locked,
	).
		Scan(&id)
	if err != nil {
		log.Print("111")
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ID := CreateResponse{
		Id: id,
	}

	// Отправляем успешный статус
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ID)
}

func getUserById(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	var user UserToGet

	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM users WHERE id = $1", name, surname, email, avatar, login, password, role, weight, height, locked)

	err = db.QueryRow(query, id).
		Scan(
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Avatar,
			&user.Login,
			&user.Password,
			&user.Role,
			&user.Weight,
			&user.Height,
			&user.Locked,
		)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	user.Id = id

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

}

func updateUserById(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Декодируем JSON из тела запроса в структуру User
	var updateUser User
	err = json.NewDecoder(r.Body).Decode(&updateUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Формируем запрос к базе данных для обновления пользователя
	query := fmt.Sprintf("UPDATE users SET %s=$1, %s=$2, %s=$3, %s=$4, %s=$5, %s=$6, %s=$7, %s=$8, %s=$9, %s=$10 WHERE id=$11", name, surname, email, avatar, login, password, role, weight, height, locked)
	_, err = db.Exec(query,
		updateUser.Name, updateUser.Surname, updateUser.Email, updateUser.Avatar, updateUser.Login, updateUser.Password,
		updateUser.Role, updateUser.Weight, updateUser.Height, updateUser.Locked, id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Отправляем успешный статус
	w.WriteHeader(http.StatusOK)
}

func deleteUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Удаляем пользователя из базы данных
	result, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Проверяем, сколько строк было затронуто удалением
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Если ни одна строка не была затронута удалением, значит пользователь с указанным id не найден
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Отправляем успешный статус, если пользователь успешно удален
	w.WriteHeader(http.StatusOK)
}

func getId(r *http.Request) (int64, error) {
	id := r.URL.Path[len("/users/"):]
	idRes, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, err
	}

	return idRes, nil
}
