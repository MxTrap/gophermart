package posgress

const insertStmt = "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"

const findByIdStmt = "SELECT * FROM users WHERE id = $1"

const findByUsernameStmt = "SELECT * FROM users WHERE login = $1"
