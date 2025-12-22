# ToDo Beginner

A simple ToDo web application built with Go.
This project is made for learning purposes by a programming beginner.

## Features

- User login / logout (session-based)
- Add new tasks
- Edit existing tasks
- Delete tasks
- Mark tasks as done / undone
- Set due date (date only)
- Set duration (minutes)
- Set cost (yen)
- Data persistence using PostgreSQL

## Why I made this

I started learning programming with no prior experience.

At first, even reading simple code was difficult.
Instead of aiming for perfect code, I focused on **building a working application step by step**.

By using AI as a learning partner, I tried to:
<<<<<<< HEAD

=======

> > > > > > > 23d3253794e97fb19f1d83fb24dc7a72e199b1e9

- understand what each part of the code does
- add features one by one
- finish the project instead of abandoning it halfway

This app represents my learning process.
<<<<<<< HEAD
=======
This project also helped me understand how backend services are deployed and operated in a real environment.

> > > > > > > 23d3253794e97fb19f1d83fb24dc7a72e199b1e9

## Tech Stack

- Language: Go
- Web: net/http, html/template
- Database: PostgreSQL
- Frontend: HTML / CSS (no framework)
- Deployment: Render
- Version Control: GitHub

## Design / Learning Points

- Session management using cookies
- CRUD operations with PostgreSQL
- Separation of concerns:
  - DB logic
  - HTTP handlers
  - Template rendering
- Formatting display data in Go instead of HTML
- Environment variable support for local and production use
  <<<<<<< HEAD
  =======
- Handling deployment-related issues such as environment variables and database connectivity

> > > > > > > 23d3253794e97fb19f1d83fb24dc7a72e199b1e9

## How to Run (Local)

1. Prepare PostgreSQL and create a database
2. Set DATABASE_URL environment variable
3. Run the app:go run main.go
4. Then open:http://localhost:8080

## Future Plans

<<<<<<< HEAD

=======

> > > > > > > 23d3253794e97fb19f1d83fb24dc7a72e199b1e9

- Improve UI design
- Add task priority
- Add multiple user support
- Add confirmation dialogs for delete
