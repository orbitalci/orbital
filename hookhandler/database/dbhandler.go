package database

import (
    "os"
    "fmt"
    "log"
    "time"
    "database/sql"

    _ "github.com/lib/pq"
    // for pretty printing objects:
    // "github.com/davecgh/go-spew/spew"
)


func PullWebhookFromPostgres() *sql.Rows {
    db, con_err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if con_err != nil {
        log.Fatal(con_err)
    }
    rows, query_err := db.Query("select * from webhook_data")
    if query_err != nil {
        log.Fatal(query_err)
    }
    return rows
}

func WriteWebhookString(repo_url string, hash string, hook_time time.Time) string {
    whitespc := "------------------------------------------"
    return fmt.Sprintf("%s\nUrl: %s \nHash: %s\nTime: %s \n\n\n", whitespc, repo_url, hash, hook_time)
}

func AddToPostgres(repourl string, git_hash string) error {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    defer db.Close()
    if err != nil {
        log.Printf("POSTGRES ERROR!!!! %s\n", err)
        return err
    }
    hook_time := time.Now().Format(time.RFC3339)
    _, err1 := db.Exec("insert into \"webhook_data\" values ($1,$2,$3)", repourl, git_hash, hook_time)
    if err1 != nil {
        log.Printf("Error!!\n")
        log.Fatal(err1)
        return err1
    }
    return nil
}

