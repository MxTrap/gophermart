package order

const selectAllStmt = "SELECT * FROM orders WHERE user_id = $1;"

const selectByNumber = `SELECT o.user_id, o.number, s.status, o.accrual, o.uploaded_at
FROM orders AS o JOIN order_statuses AS s ON o.status_id = s.id
WHERE number = $1;`

const insertStmt = `INSERT INTO orders (user_id, number, status_id, accrual, uploaded_at)
VALUES ($1,$2,
(SELECT id FROM order_statuses WHERE status=$3),
$4, $5);`

const updateStmt = `UPDATE orders
SET status_id = (SELECT id FROM order_statuses WHERE status=$1), accrual = $2
WHERE number=$3;`
