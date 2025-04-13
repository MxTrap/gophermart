package order

const selectAllStmt = "SELECT * FROM orders WHERE user_id = $1"

const selectByNumber = "SELECT number FROM orders WHERE number = $1"
