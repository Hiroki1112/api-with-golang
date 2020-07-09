package main

import (
	"database/sql"
	"encoding/json"
    "log"
	"net/http"
	//"reflect"
	//"strconv"
	"fmt"
	"os"

	"github.com/gorilla/mux"//ルーティング用ライブラリ
	"github.com/lib/pq"
    "github.com/subosito/gotenv"
)

var db *sql.DB

func init() {
	//作成した.envファイルを読み込む
	gotenv.Load()
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Article struct {
	ID int `json:id` 
	Title string `json:title`
	Author string `json:author`
	PostDate string `json:year`
}

var articles []Article//this is slice

func getArticles(w http.ResponseWriter, r *http.Request) {
	//log.Printf("Get all articles")
	var article Article
    articles = []Article{}

    rows, err := db.Query("SELECT * FROM Article;")
    logFatal(err)

    defer rows.Close()

    for rows.Next() {
        err := rows.Scan(&article.ID, &article.Title, &article.Author, &article.PostDate)
        logFatal(err)

        articles = append(articles, article)
    }
    json.NewEncoder(w).Encode(articles)
}

func getArticle(w http.ResponseWriter, r *http.Request) {
	log.Println("Get article is called")
	
	var article Article
    params := mux.Vars(r)

    rows := db.QueryRow("SELECT * FROM ARTICLE WHERE id=$1", params["id"])

    err := rows.Scan(&article.ID, &article.Title, &article.Author, &article.PostDate)
    logFatal(err)

    json.NewEncoder(w).Encode(article)
}

func addArticle(w http.ResponseWriter, r *http.Request) {
	//log.Println("Add article is called")
	var article Article
	var articleID int

	// json -> struct ?
	json.NewDecoder(r.Body).Decode(&article)
	
	err := db.QueryRow("INSERT INTO ARTICLE (title, author, postdate) values($1, $2, $3) RETURNING id;",
		article.Title, article.Author, article.PostDate).Scan(&articleID)
		
	logFatal(err)
	
    json.NewEncoder(w).Encode(articles)
}

func updateArticle(w http.ResponseWriter, r *http.Request) {
	log.Println("Update article is called")
	
	var article Article
	json.NewDecoder(r.Body).Decode(&article)

	for i, item := range articles {
		if item.ID == article.ID {
			articles[i] = article
		}
	}

	json.NewEncoder(w).Encode(article)
}

func removeArticle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

    result, err := db.Exec("DELETE FROM ARTICLE WHERE id=$1", params["id"])
    logFatal(err)

    rowsDeleted, err := result.RowsAffected()
    logFatal(err)

    fmt.Println("rowsDeleted", rowsDeleted)
    json.NewEncoder(w).Encode(rowsDeleted)
}


func main() {

	pgURL, err := pq.ParseURL(os.Getenv("ELEPHANTSQL_URL"))
    logFatal(err)
    log.Println("pgUrl: ", pgURL)

    // Connect to postgres
    db, err = sql.Open("postgres", pgURL)
    logFatal(err)

    err = db.Ping()
    logFatal(err)

	// ルーターを作成。リクエストを捌く。https://github.com/gorilla/mux
	router := mux.NewRouter()

	/*articles = append(articles,
        Article{ID: 1, Title: "Article1", Author: "Gopher", PostDate: "2019/1/1"},
        Article{ID: 2, Title: "Article2", Author: "Gopher", PostDate: "2019/2/2"},
        Article{ID: 3, Title: "Article3", Author: "Gopher", PostDate: "2019/3/3"},
        Article{ID: 4, Title: "Article4", Author: "Gopher", PostDate: "2019/4/4"},
        Article{ID: 5, Title: "Article5", Author: "Gopher", PostDate: "2019/5/5"},
    )*/

	//エンドポイント,ハンドラ？
	router.HandleFunc("/articles",getArticles).Methods("GET")
	router.HandleFunc("/articles/{id}",getArticle).Methods("GET")
	router.HandleFunc("/articles",addArticle).Methods("POST")
	router.HandleFunc("/articles",updateArticle).Methods("PUT")
	router.HandleFunc("/articles/{id}",removeArticle).Methods("DELETE")
	

	//サーバーを立ち上げる
	log.Println("Listen server now.....")
	//異常があった場合、処理を停止する
	log.Fatal(http.ListenAndServe(":8080",router))

}