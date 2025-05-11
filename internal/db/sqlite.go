package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/YattaDeSune/calc-project/internal/entities"
	"github.com/YattaDeSune/calc-project/internal/errors"
	"github.com/YattaDeSune/calc-project/internal/logger"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type Database struct {
	db     *sql.DB
	logger *zap.Logger
}

func New(ctx context.Context) (*Database, error) {
	logger := logger.FromContext(ctx)

	db, err := sql.Open("sqlite3", "calculator.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		db:     db,
		logger: logger,
	}

	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

func (d *Database) createTables() error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		login TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);`

	expressionsTable := `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		expression TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		result NUMERIC,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id)
	);`

	if _, err := d.db.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if _, err := d.db.Exec(expressionsTable); err != nil {
		return fmt.Errorf("failed to create expressions table: %w", err)
	}

	d.logger.Info("Database tables created successfully")
	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) CreateUser(ctx context.Context, login, password string) (int, error) {
	existingUser, err := d.GetUserByLogin(ctx, login)
	if err != nil && err != errors.ErrWrongLogin {
		return 0, fmt.Errorf("failed to check user existence: %w", err)
	}

	if existingUser != nil {
		return 0, errors.ErrUserExists
	}

	const query = `INSERT INTO users (login, password) VALUES (?, ?)`
	result, err := d.db.ExecContext(ctx, query, login, password)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(id), nil
}

func (d *Database) GetUserByLogin(ctx context.Context, login string) (*entities.User, error) {
	const query = `SELECT id, login, password FROM users WHERE login = ?`
	row := d.db.QueryRowContext(ctx, query, login)

	var user entities.User
	if err := row.Scan(&user.ID, &user.Login, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrWrongLogin
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &user, nil
}

func (d *Database) CreateExpression(ctx context.Context, expr string, userID int, status string) (int, error) {
	const query = `INSERT INTO expressions (expression, user_id, status) VALUES (?, ?, ?)`
	result, err := d.db.ExecContext(ctx, query, expr, userID, status)
	if err != nil {
		return 0, fmt.Errorf("failed to create expression: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(id), nil
}

func (d *Database) GetExpressionByID(ctx context.Context, id int, userID int) (*entities.ExpressionDB, error) {
	const query = `
	SELECT id, expression, user_id, status, result, created_at FROM expressions 
	WHERE id = ?
	AND user_id = ?
	`
	row := d.db.QueryRowContext(ctx, query, id, userID)

	var expr entities.ExpressionDB
	if err := row.Scan(&expr.ID, &expr.Expression, &expr.UserID, &expr.Status, &expr.Result, &expr.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan expression: %w", err)
	}

	return &expr, nil
}

func (d *Database) GetExpressionsByUser(ctx context.Context, userID int) ([]entities.ExpressionDB, error) {
	const query = `SELECT id, expression, user_id, status, result, created_at FROM expressions WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := d.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query expressions: %w", err)
	}
	defer rows.Close()

	var expressions []entities.ExpressionDB
	for rows.Next() {
		var expr entities.ExpressionDB
		if err := rows.Scan(&expr.ID, &expr.Expression, &expr.UserID, &expr.Status, &expr.Result, &expr.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan expression: %w", err)
		}
		expressions = append(expressions, expr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return expressions, nil
}

func (d *Database) UpdateExpressionStatus(ctx context.Context, id int, status string) error {
	const query = `UPDATE expressions SET status = ? WHERE id = ?`
	_, err := d.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update expression status: %w", err)
	}
	return nil
}

func (d *Database) UpdateExpressionResult(ctx context.Context, id int, result any, status string) error {
	const query = `UPDATE expressions SET result = ?, status = ? WHERE id = ?`
	_, err := d.db.ExecContext(ctx, query, result, status, id)
	if err != nil {
		return fmt.Errorf("failed to update expression result: %w", err)
	}
	return nil
}
