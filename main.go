package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// ---------------------
// データ構造
// ---------------------

type User struct {
	Name     string
	Password string
}

type Todo struct {
	Task        string
	Due         string
	Duration    int
	Cost        int
	DurationStr string
	CostStr     string
	Done        bool
}

var users = []User{
	{Name: "admin", Password: "1234"},
}

const dataDir = "data"

// ---------------------
// メインのハンドラー
// ---------------------

func handler(w http.ResponseWriter, r *http.Request) {

	loggedIn := false
	username := ""
	todos := []Todo{}

	// ● Cookie でログイン判定
	cookie, err := r.Cookie("session_user")
	if err == nil && cookie.Value != "" {
		username = cookie.Value
		loggedIn = true
		todos = loadTodosForUser(username)
	}

	// ---------------------
	// ログイン処理
	// ---------------------
	if r.Method == "POST" && r.FormValue("login") != "" {
		name := r.FormValue("username")
		pass := r.FormValue("password")

		for _, u := range users {
			if u.Name == name && u.Password == pass {

				// Cookie 保存
				http.SetCookie(w, &http.Cookie{
					Name:  "session_user",
					Value: name,
					Path:  "/",
				})

				loggedIn = true
				username = name
				todos = loadTodosForUser(name)

				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
	}

	// ログインしていない場合 → login.html へ
	if !loggedIn {
		renderTemplate(w, "login.html", nil)
		return
	}

	// ---------------------
	// ログアウト
	// ---------------------
	if r.Method == "POST" && r.FormValue("logout") != "" {
		http.SetCookie(w, &http.Cookie{
			Name:   "session_user",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// ---------------------
	// 完了トグル
	// ---------------------
	if r.Method == "POST" && r.FormValue("toggle") != "" {
		i, _ := strconv.Atoi(r.FormValue("toggle"))
		if i >= 0 && i < len(todos) {
			todos[i].Done = !todos[i].Done
			saveTodosForUser(username, todos)
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// ---------------------
	// 新規追加
	// ---------------------
	if r.Method == "POST" && r.FormValue("task") != "" {
		task := r.FormValue("task")
		due := r.FormValue("due")
		duration, _ := strconv.Atoi(r.FormValue("duration"))
		cost, _ := strconv.Atoi(r.FormValue("cost"))

		todos = append(todos, Todo{
			Task:        task,
			Due:         due,
			Duration:    duration,
			Cost:        cost,
			DurationStr: strconv.Itoa(duration) + "分",
			CostStr:     strconv.Itoa(cost) + "円",
			Done:        false,
		})

		saveTodosForUser(username, todos)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// ---------------------
	// 編集フォーム表示
	// ---------------------
	if r.URL.Path == "/edit" && r.Method == "GET" {

		id, _ := strconv.Atoi(r.URL.Query().Get("id"))

		if id < 0 || id >= len(todos) {
			http.NotFound(w, r)
			return
		}

		renderTemplate(w, "edit.html", struct {
			ID   int
			Todo Todo
		}{ID: id, Todo: todos[id]})

		return
	}

	// ---------------------
	// 編集保存処理
	// ---------------------
	if r.URL.Path == "/edit" && r.Method == "POST" {

		id, _ := strconv.Atoi(r.FormValue("id"))

		if id >= 0 && id < len(todos) {

			todos[id].Task = r.FormValue("task")
			todos[id].Due = r.FormValue("due")

			duration, _ := strconv.Atoi(r.FormValue("duration"))
			cost, _ := strconv.Atoi(r.FormValue("cost"))

			todos[id].Duration = duration
			todos[id].Cost = cost

			todos[id].DurationStr = strconv.Itoa(duration) + "分"
			todos[id].CostStr = strconv.Itoa(cost) + "円"

			saveTodosForUser(username, todos)
		}

		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// ---------------------
	// 一覧ページ表示
	// ---------------------
	renderTemplate(w, "tasks.html", struct {
		UserName string
		Todos    []Todo
	}{
		UserName: username,
		Todos:    todos,
	})
}

// ---------------------
// テンプレート描画
// ---------------------
func renderTemplate(w http.ResponseWriter, file string, data interface{}) {
	t := template.Must(template.ParseFiles(
		"templates/login.html",
		"templates/tasks.html",
		"templates/edit.html",
	))
	t.ExecuteTemplate(w, file, data)
}

// ---------------------
// JSON 保存
// ---------------------
func saveTodosForUser(username string, todos []Todo) {
	os.MkdirAll(dataDir, 0755)
	filename := dataDir + "/" + username + ".json"

	data, _ := json.MarshalIndent(todos, "", "  ")
	ioutil.WriteFile(filename, data, 0644)
}

// ---------------------
// JSON 読み込み
// ---------------------
func loadTodosForUser(username string) []Todo {
	filename := dataDir + "/" + username + ".json"

	if _, err := os.Stat(filename); err == nil {
		data, _ := ioutil.ReadFile(filename)
		var t []Todo
		json.Unmarshal(data, &t)
		return t
	}

	return []Todo{}
}

// ---------------------
func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/edit", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
}
