package withdrawn

const selectStmt = "SELECT number, sum, processed_at FROM withdrawals WHERE user_id = $1;"

const insertStmt = `INSERT INTO withdrawals(user_id, number, sum, processed_at) VALUES ($1, $2, $3, $4);`
