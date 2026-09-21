package main

import (
	"database/sql"
	"log"
	"net/http"
	"flag"
	"os"
	"snippetbox.alexewards.net/internal/models"
	_ "github.com/go-sql-driver/mysql"

	"html/template"
)

type application struct {	
	infoLog *log.Logger
	errLog  *log.Logger
	snippets *models.SnippetModel
	templateCache map[string]*template.Template
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)	
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP network address")

	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}
	a := &application{
		infoLog: infoLog,
		errLog:  errLog,
		snippets: &models.SnippetModel{DB: db},
		templateCache: templateCache,}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// mux.Handle("/", http.HandlerFunc(home))
	// mux.Handle("/snippet/view", http.HandlerFunc(snippetView))
	// mux.Handle("/snippet/create", http.HandlerFunc(snippetCreate))

	mux.HandleFunc("/", a.home)
	mux.HandleFunc("/snippet/view", a.snippetView)
	mux.HandleFunc("/snippet/create", a.snippetCreate)	

	
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: a.errLog,
		Handler:  mux,
	}

	a.infoLog.Println("Server starting on", *addr)	
	err = srv.ListenAndServe()

	if err != nil {
		a.errLog.Fatal(err)
	}
}
