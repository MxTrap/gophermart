package balance

const increaseBalanceStmt = `UPDATE users SET balance = balance + $1 WHERE id = $2;`

const withdrawalStmt = `UPDATE users SET balance = balance - $1, withdrawn=withdrawn + $1 WHERE id = $2;`

const selectStmt = `SELECT balance, withdrawn FROM users WHERE id = $1`
