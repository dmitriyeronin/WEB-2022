package main

import (
  "context"
  "fmt"
  "time"
  "net/http"
  "html/template"
  "encoding/json"

  "go.uber.org/zap"

  "github.com/go-redis/redis"
  kafka "github.com/segmentio/kafka-go"
)

var client *redis.Client

var logger *zap.Logger
var sugar *zap.SugaredLogger

func track(msg string) (string, time.Time) {
    return msg, time.Now()
}

func duration(msg string, start time.Time) {
    sugar.Info(fmt.Sprintf("%v: %v\n", msg, time.Since(start)))
}

type Article struct {
  Time, Title, Text string
}

var posts = []Article{}

func main_page(w http.ResponseWriter, r *http.Request) {
  defer duration(track("main_page"))

  t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
  if err != nil {
    sugar.Error(err)
  }

  //db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:33066)/HW4")
  //if err != nil {
  //  sugar.Errorln(err)
  //}
  //defer db.Close()

  //res, err := db.Query("select * from `articles`")
  //if err != nil {
  //  sugar.Error(err)
  //}

  //posts = []Article{}
  //for res.Next() {
  //  var post Article
  //  err = res.Scan(&post.Id, &post.Title, &post.Text)
  //  if err != nil {
  //    sugar.Error(err)
  //  }

  //  posts = append(posts, post)

  //}
  //posts = nil
  //t.ExecuteTemplate(w, "index", posts)

  cmd := redis.NewStringCmd("select", 0)
	err = client.Process(cmd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var post Article

  posts = []Article{}

  keys := client.Keys("*")
  for _, key := range keys.Val() {
		val, err := client.Get(key).Result()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
    err = json.Unmarshal([]byte(val), &post)
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
    data := map[string]interface{}{
		  "title": title,
      "text": article_text,
	  }

    msg, err := json.Marshal(data)
	  if err != nil {
		  sugar.Error(err)
	  }

    conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "Articles", 0)
  	if err != nil {
  		sugar.Error(err)
  	}
  	defer conn.Close()

    err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			sugar.Error(err)
		}

    _, err = conn.WriteMessages(
			kafka.Message{Value: msg},
		)
		if err != nil {
			sugar.Error(err)
		}

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

  client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

  logger.Info("Server started.")

  HandleFunc()
}
