package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// Todo構造体
type Todo struct {
	ID           int
	Task         string
	Due          string
	Duration     int
	Cost         int
	Done         bool
	DueFormatted string
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

	// ------------------------------------------------------------
	// 📌 ユーザーテーブル作成
	// ------------------------------------------------------------
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	)
`)
	if err != nil {
		log.Fatal("users テーブル作成エラー:", err)
	}

	// ------------------------------------------------------------
	// 📌 初期ユーザー（admin / 1234）を1回だけ作成
	// ------------------------------------------------------------
	_, err = db.Exec(`
	INSERT INTO users (username, password)
	VALUES ('admin', '1234')
	ON CONFLICT (username) DO NOTHING
`)
	if err != nil {
		log.Fatal("admin ユーザー作成エラー:", err)
	}

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
	http.HandleFunc("/login", handleLogin)
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
	// 🔒 ログインチェック
	user, ok := getLoginUser(r)
	if !ok {
		http.Redirect(w, r, "/login", 303)
		return
	}

	rows, err := db.Query(
		"SELECT id, task, due, duration, cost, done FROM todos ORDER BY id DESC",
	)
	if err != nil {
		http.Error(w, "データ取得エラー", 500)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		rows.Scan(&t.ID, &t.Task, &t.Due, &t.Duration, &t.Cost, &t.Done)

		// ★ 表示用の日付をここで作る
		t.DueFormatted = formatDate(t.Due)

		todos = append(todos, t)
	}

	// ユーザー名も一緒に渡す
	templates.ExecuteTemplate(w, "tasks.html", struct {
		UserName string
		Todos    []Todo
	}{
		UserName: user,
		Todos:    todos,
	})
}

// ------------------------------------------------------------
// 📌 セッションCookie名
// ------------------------------------------------------------
const sessionName = "todo_session"

// ------------------------------------------------------------
// 📌 ログイン中ユーザー取得
// Cookie があればログイン済みと判断する
// ------------------------------------------------------------
func getLoginUser(r *http.Request) (string, bool) {
	c, err := r.Cookie(sessionName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

// ------------------------------------------------------------
// 📌 新規追加
// ------------------------------------------------------------
func handleAdd(w http.ResponseWriter, r *http.Request) {
	// 🔒 ログイン必須
	_, ok := getLoginUser(r)
	if !ok {
		http.Redirect(w, r, "/login", 303)
		return
	}

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
	// 🔒 ログイン必須
	_, ok := getLoginUser(r)
	if !ok {
		http.Redirect(w, r, "/login", 303)
		return
	}

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
	// 🔒 ログイン必須
	_, ok := getLoginUser(r)
	if !ok {
		http.Redirect(w, r, "/login", 303)
		return
	}

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
	// 🔒 ログイン必須
	_, ok := getLoginUser(r)
	if !ok {
		http.Redirect(w, r, "/login", 303)
		return
	}

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
	// 🔒 ログイン必須
	_, ok := getLoginUser(r)
	if !ok {
		http.Redirect(w, r, "/login", 303)
		return
	}

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
// 📌 ログイン処理
// ------------------------------------------------------------
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username=$1 AND password=$2",
		username, password,
	).Scan(&count)

	if err != nil || count == 0 {
		http.Error(w, "ログイン失敗", 401)
		return
	}

	// ログイン成功 → Cookie 発行
	http.SetCookie(w, &http.Cookie{
		Name:  sessionName,
		Value: username,
		Path:  "/",
	})

	http.Redirect(w, r, "/", 303)
}

// ------------------------------------------------------------
// 📌 ログアウト
// Cookie を削除する
// ------------------------------------------------------------
func handleLogout(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:   sessionName,
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Cookie 削除
	})
	http.Redirect(w, r, "/login", 303)
}

// ------------------------------------------------------------
// 📌 日付を YYYY-MM-DD 形式に整形
// DBの値が 2025-12-12T00:00:00Z などでも安全
// ------------------------------------------------------------
func formatDate(d string) string {
	if d == "" {
		return ""
	}

	// PostgreSQL / HTML date 両対応
	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, d); err == nil {
			return t.Format("2006-01-02")
		}
	}

	// パースできなければ元の文字列を返す（保険）
	return d
}
