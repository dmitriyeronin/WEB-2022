package main

import (
  "fmt"
  "time"
  "net/http"
  "html/template"
  "database/sql"

  "go.uber.org/zap"

  _ "github.com/go-sql-driver/mysql"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger

func track(msg string) (string, time.Time) {
    return msg, time.Now()
}

func duration(msg string, start time.Time) {
    sugar.Info(fmt.Sprintf("%v: %v\n", msg, time.Since(start)))
}

type Article struct {
  Id uint16
  Title, Text string
}

var posts = []Article{}

func main_page(w http.ResponseWriter, r *http.Request) {
  defer duration(track("main_page"))

  t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")

  if err != nil {
    sugar.Error(err)
  }

  db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/HW4")
  if err != nil {
    sugar.Errorln(err)
  }
  defer db.Close()

  res, err := db.Query("select * from `articles`")
  if err != nil {
    sugar.Error(err)
  }

  posts = []Article{}
  for res.Next() {
    var post Article
    err = res.Scan(&post.Id, &post.Title, &post.Text)
    if err != nil {
      sugar.Error(err)
    }

    posts = append(posts, post)

  }

  t.ExecuteTemplate(w, "index", posts)
}

func new_article(w http.ResponseWriter, r *http.Request)  {
  t, err := template.ParseFiles("templates/new_article.html", "templates/header.html", "templates/footer.html")

  if err != nil {
    sugar.Error(err)
  }

  t.ExecuteTemplate(w, "new_article", nil)
}

func save_article(w http.ResponseWriter, r *http.Request)  {
  defer duration(track("save_article"))
  title := r.FormValue("title")
  article_text := r.FormValue("article_text")

  if title == "" || article_text == "" {
    fmt.Fprintf(w, "Введите название и текс статьи.")
  } else {
    db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/HW4")
    if err != nil {
      sugar.Error(err)
    }
    defer db.Close()

    insert, err := db.Query(fmt.Sprintf("insert into `articles` (`title`,`article_text`) values('%s', '%s')", title, article_text))
    if err != nil {
      sugar.Error(err)
    }
    defer insert.Close()

    http.Redirect(w, r, "/", 301)
  }
}

func HandleFunc() {
  http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
  http.HandleFunc("/", main_page)
  http.HandleFunc("/new_article", new_article)
  http.HandleFunc("/save_article", save_article)
  http.ListenAndServe(":8080", nil)
}

func main() {
  logger, _ = zap.NewProduction()
	defer logger.Sync()
	sugar = logger.Sugar()

  logger.Info("Server started.")

  HandleFunc()
}
