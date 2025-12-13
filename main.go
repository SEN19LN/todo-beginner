package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

// Todo構造体
type Todo struct {
	ID       int
	Task     string
	Due      string
	Duration int
	Cost     int
	Done     bool
}

// DB変数（グローバル）
var db *sql.DB

// ------------------------------------------------------------
// 📌 DBに接続（PostgreSQL）
// ------------------------------------------------------------
func initDB() {
	var err error

	// Render とローカル両対応
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// ローカル開発用
		dsn = "postgres://postgres:tkhr0719@localhost:5432/todoapp?sslmode=disable"
	}

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("DB接続エラー:", err)
	}

	// 接続テスト
	if err = db.Ping(); err != nil {
		log.Fatal("DBが起動していません:", err)
	}

	// テーブル作成
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			task TEXT NOT NULL,
			due TEXT,
			duration INT,
			cost INT,
			done BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		log.Fatal("テーブル作成エラー:", err)
	}
	log.Println("DB準備完了")
}

// ------------------------------------------------------------
// 📌 テンプレートロード
// ------------------------------------------------------------
var templates = template.Must(template.ParseGlob("templates/*.html"))

// ------------------------------------------------------------
// 📌 ルーティング設定
// ------------------------------------------------------------
func main() {
	initDB() // DB初期化

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/toggle", handleToggle)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/edit", handleEditPage)
	http.HandleFunc("/update", handleUpdate)
	http.HandleFunc("/logout", handleLogout)

	// PORT は Render が自動設定 → ローカルは 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("起動中 http://localhost:" + port)
	http.ListenAndServe(":"+port, nil)
}

// ------------------------------------------------------------
// 📌 トップページ（一覧表示）
// ------------------------------------------------------------
func handleIndex(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, task, due, duration, cost, done FROM todos ORDER BY id DESC")
	if err != nil {
		http.Error(w, "データ取得エラー", 500)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		rows.Scan(&t.ID, &t.Task, &t.Due, &t.Duration, &t.Cost, &t.Done)
		todos = append(todos, t)
	}

	// ★ テンプレートが期待する形（UserName + Todos）
	data := struct {
		UserName string
		Todos    []Todo
	}{
		UserName: "admin", // 今は固定（後でログインユーザー名を入れる）
		Todos:    todos,
	}

	templates.ExecuteTemplate(w, "tasks.html", data)
}

// ------------------------------------------------------------
// 📌 新規追加
// ------------------------------------------------------------
func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	task := r.FormValue("task")
	due := r.FormValue("due")
	dur, _ := strconv.Atoi(r.FormValue("duration"))
	cost, _ := strconv.Atoi(r.FormValue("cost"))

	_, err := db.Exec(
		"INSERT INTO todos (task, due, duration, cost) VALUES ($1, $2, $3, $4)",
		task, due, dur, cost,
	)
	if err != nil {
		http.Error(w, "追加エラー", 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ------------------------------------------------------------
// 📌 完了トグル（Done <-> 未完了）
// ------------------------------------------------------------
func handleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 303)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Redirect(w, r, "/", 303)
		return
	}

	_, err := db.Exec("UPDATE todos SET done = NOT done WHERE id = $1", id)
	if err != nil {
		http.Error(w, "更新エラー", 500)
		return
	}

	http.Redirect(w, r, "/", 303)
}

// ------------------------------------------------------------
// 📌 削除
// ------------------------------------------------------------
func handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.Redirect(w, r, "/", 303)
		return
	}

	_, err := db.Exec("DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		http.Error(w, "削除エラー", 500)
		return
	}

	http.Redirect(w, r, "/", 303)
}

// ------------------------------------------------------------
// 📌 編集ページの表示（edit.html）
// ------------------------------------------------------------
func handleEditPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var t Todo
	err := db.QueryRow(
		"SELECT id, task, due, duration, cost, done FROM todos WHERE id = $1",
		id,
	).Scan(&t.ID, &t.Task, &t.Due, &t.Duration, &t.Cost, &t.Done)

	if err != nil {
		http.Error(w, "データ取得エラー", 500)
		return
	}

	templates.ExecuteTemplate(w, "edit.html", t)
}

// ------------------------------------------------------------
// 📌 編集内容の保存
// ------------------------------------------------------------
func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 303)
		return
	}

	id := r.FormValue("id")
	task := r.FormValue("task")
	due := r.FormValue("due")
	dur, _ := strconv.Atoi(r.FormValue("duration"))
	cost, _ := strconv.Atoi(r.FormValue("cost"))

	_, err := db.Exec(
		"UPDATE todos SET task = $1, due = $2, duration = $3, cost = $4 WHERE id = $5",
		task, due, dur, cost, id,
	)
	if err != nil {
		http.Error(w, "更新エラー", 500)
		return
	}

	http.Redirect(w, r, "/", 303)
}

// ------------------------------------------------------------
// 📌 ログアウト（※現状はログインなしなのでダミー）
// ------------------------------------------------------------
func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 303)
}
