package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type Todo struct {
	Task        string
	Due         string
	Duration    int
	Cost        int
	DurationStr string
	CostStr     string
	Done        bool
}

var todos []Todo

const dataFile = "todo.json"

func saveTodos() {
	data, _ := json.MarshalIndent(todos, "", "  ")
	_ = ioutil.WriteFile(dataFile, data, 0644)
}

func loadTodos() {
	if _, err := os.Stat(dataFile); err == nil {
		data, _ := ioutil.ReadFile(dataFile)
		_ = json.Unmarshal(data, &todos)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// 完了トグル処理
	if r.Method == "POST" && r.FormValue("toggle") != "" {
		i, _ := strconv.Atoi(r.FormValue("toggle"))
		if i >= 0 && i < len(todos) {
			todos[i].Done = !todos[i].Done
			saveTodos()

		}
	}

	// 新規追加処理
	if r.Method == "POST" && r.FormValue("task") != "" {
		task := r.FormValue("task")
		due := r.FormValue("due")
		durationInput, _ := strconv.Atoi(r.FormValue("duration"))
		costInput, _ := strconv.Atoi(r.FormValue("cost"))

		durationStr := ""
		costStr := ""

		if durationInput == 15 {
			durationStr = "15分以内"
		} else {
			durationStr = strconv.Itoa(durationInput) + "分"
		}

		if costInput == 1000 {
			costStr = "1000円以内"
		} else {
			costStr = strconv.Itoa(costInput) + "円"
		}

		todos = append(todos, Todo{
			Task:        task,
			Due:         due,
			Duration:    durationInput,
			Cost:        costInput,
			DurationStr: durationStr,
			CostStr:     costStr,
			Done:        false,
		})
		saveTodos()

	}

	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>ToDo Beginner</title>
		<style>
	body {
		font-family: sans-serif;
		background: #f5f5f5;
		padding: 20px;
	}

	h1 {
		text-align: center;
	}

	form {
		background: white;
		padding: 15px;
		border-radius: 8px;
		margin-bottom: 20px;
		box-shadow: 0 0 5px rgba(0,0,0,0.1);
	}

	input, button {
		font-size: 16px;
		padding: 8px;
		margin: 5px 0;
		width: 100%;
		box-sizing: border-box;
	}

	button {
		background: #007bff;
		color: white;
		border: none;
		border-radius: 5px;
		cursor: pointer;
	}

	button:hover {
		background: #0056b3;
	}

	.done {
		text-decoration: line-through;
		color: gray;
	}

	/* 数値入力の上下ボタンを消す */
	input[type=number]::-webkit-inner-spin-button,
	input[type=number]::-webkit-outer-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}

	input[type=number] {
		-moz-appearance: textfield;
	}

	ul {
		list-style: none;
		padding: 0;
	}

	li {
		background: white;
		padding: 10px;
		margin-bottom: 10px;
		border-radius: 5px;
	}

	.check-btn {
		width: auto;
		padding: 4px 8px;
		font-size: 14px;
		background: #ddd;
		color: black;
		margin-right: 8px;
	}
</style>


	</head>
	<body>
		<h1>📝 ToDo Beginner</h1>

		<form method="POST">
			<input type="text" name="task" placeholder="やること">
			<input type="date" name="due">
			<input type="number" name="duration" placeholder="所要時間（分）" step="1">
			<input type="number" name="cost" placeholder="費用（円）" step="1">
			<button type="submit">追加</button>
		</form>

		<ul>
			{{range $i, $t := .}}
				<li>
					<form method="POST" style="display:inline;">
						<button class="check-btn" name="toggle" value="{{$i}}">✅</button>
					</form>

					<span class="{{if $t.Done}}done{{end}}">
						{{$t.Task}}
						（期限：{{$t.Due}} /
						時間：{{$t.DurationStr}} /
						費用：{{$t.CostStr}}）
					</span>
				</li>
			{{end}}
		</ul>
	</body>
	</html>
	`

	t, _ := template.New("web").Parse(tmpl)
	t.Execute(w, todos)
}

func main() {
	loadTodos()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
