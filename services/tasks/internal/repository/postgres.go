package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tech-ip-sem2/services/tasks/internal/models"

	_ "github.com/lib/pq"
)

type TaskRepository interface {
	Create(task models.Task) (models.Task, error)
	GetAll() ([]models.Task, error)
	GetByID(id string) (models.Task, error)
	Update(id string, task models.Task) (models.Task, error)
	Delete(id string) error
	SearchByTitle(keyword string) ([]models.Task, error)
	SearchByTitleUnsafe(keyword string) ([]models.Task, error)
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(host, port, user, password, dbname string) (*PostgresRepo, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// Create — параметризованный запрос
func (r *PostgresRepo) Create(task models.Task) (models.Task, error) {
	id := fmt.Sprintf("t_%d", time.Now().UnixNano())
	task.ID = id
	task.CreatedAt = time.Now()

	query := `
		INSERT INTO tasks (id, title, description, due_date, done, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(query,
		task.ID, task.Title, task.Description, task.DueDate, task.Done, task.CreatedAt,
	)
	if err != nil {
		return models.Task{}, err
	}

	return task, nil
}

// GetAll — параметризованный запрос
func (r *PostgresRepo) GetAll() ([]models.Task, error) {
	query := `SELECT id, title, description, due_date, done, created_at FROM tasks ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// GetByID — параметризованный запрос
func (r *PostgresRepo) GetByID(id string) (models.Task, error) {
	query := `SELECT id, title, description, due_date, done, created_at FROM tasks WHERE id = $1`

	var t models.Task
	err := r.db.QueryRow(query, id).Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return models.Task{}, nil
	}
	if err != nil {
		return models.Task{}, err
	}

	return t, nil
}

// Update — параметризованный запрос
func (r *PostgresRepo) Update(id string, task models.Task) (models.Task, error) {
	// Сначала получаем существующую
	existing, err := r.GetByID(id)
	if err != nil || existing.ID == "" {
		return models.Task{}, err
	}

	if task.Title != "" {
		existing.Title = task.Title
	}
	if task.Description != "" {
		existing.Description = task.Description
	}
	if task.DueDate != "" {
		existing.DueDate = task.DueDate
	}
	existing.Done = task.Done

	query := `
		UPDATE tasks 
		SET title = $1, description = $2, due_date = $3, done = $4
		WHERE id = $5
	`

	_, err = r.db.Exec(query, existing.Title, existing.Description, existing.DueDate, existing.Done, id)
	if err != nil {
		return models.Task{}, err
	}

	return existing, nil
}

// Delete — параметризованный запрос
func (r *PostgresRepo) Delete(id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// SearchByTitle - БЕЗОПАСНЫЙ
func (r *PostgresRepo) SearchByTitle(keyword string) ([]models.Task, error) {
	query := `SELECT id, title, description, due_date, done, created_at 
              FROM tasks WHERE title ILIKE $1` // ILIKE - регистронезависимый поиск

	rows, err := r.db.Query(query, "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *PostgresRepo) SearchByTitleUnsafe(keyword string) ([]models.Task, error) {
	query := fmt.Sprintf("SELECT id, title, description, due_date, done, created_at FROM tasks WHERE title LIKE '%%%s%%'", keyword)

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}
