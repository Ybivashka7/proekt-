package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Article struct {
	Id       int
	Title    string
	Anons    string
	FullText string
}

var tmpl = template.Must(template.ParseFiles(
	"templates/index.html",
	"templates/create.html",
	"templates/show.html",
	"templates/header.html",
	"templates/footer.html",
))

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "qwe"
	dbname   = "Users"
)

func dbConnect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
	return sql.Open("postgres", psqlInfo)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	db, err := dbConnect()
	if err != nil {
		http.Error(w, "Ошибка бд", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, anons, full_text FROM article")
	if err != nil {
		http.Error(w, "Ошибка html "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		err = rows.Scan(&a.Id, &a.Title, &a.Anons, &a.FullText)
		if err != nil {
			log.Println(err)
			continue
		}
		articles = append(articles, a)
	}

	tmpl.ExecuteTemplate(w, "index", articles)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "create", nil)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	anons := r.FormValue("anons")
	fullText := r.FormValue("full_text")

	db, err := dbConnect()
	if err != nil {
		http.Error(w, "Ошибка подключения БД", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO article (title, anons, full_text) VALUES ($1, $2, $3)", title, anons, fullText)
	if err != nil {
		http.Error(w, "Ошибка сохранения "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/post/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	db, err := dbConnect()
	if err != nil {
		http.Error(w, "Ошибка подключения бд", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, title, anons, full_text FROM article WHERE id = $1", id)

	var a Article
	err = row.Scan(&a.Id, &a.Title, &a.Anons, &a.FullText)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tmpl.ExecuteTemplate(w, "show", a)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/save", saveHandler)
	http.HandleFunc("/post/", showPostHandler)

	fmt.Println("Сервер запущен")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

