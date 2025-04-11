package postgres

const selectSmtp = `SELECT refresh_token, expiration_time FROM tokens WHERE refresh_token = $1`

const selectTokensStmt = `SELECT id, user_id, refresh_token, expiration_time FROM tokens WHERE user_id = $1`

const insertTokenStmt = `INSERT INTO tokens (user_id, refresh_token, expiration_time) VALUES ($1, $2, $3)`

const deleteTokenStmt = `DELETE FROM tokens WHERE refresh_token = $1`

const deleteAllTokensStmt = `DELETE FROM tokens WHERE user_id = $1`
