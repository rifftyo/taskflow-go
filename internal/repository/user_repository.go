package repository

import (
	"context"
	"encoding/json"
	"errors"
	"taskflow/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users(name, email)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		user.Name,
		user.Email,
	).Scan(&user.ID, &user.CreatedAt)

	return err;
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, name, email, created_at
		FROM users
		WHERE id = $1
	`

	var user model.User

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	query := `
		SELECT id, name, email, created_at
		FROM users
		ORDER BY id DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)

	for rows.Next() {
		var user model.User

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) ListWithTasks(ctx context.Context) ([]model.UserWithTasks, error) {
	query := `
		SELECT
			u.id,
			u.name,
			u.email,
			u.created_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', t.id,
						'user_id', t.user_id,
						'title', t.title,
						'description', t.description,
						'status', t.status,
						'created_at', t.created_at,
						'updated_at', t.updated_at
					)
					ORDER BY t.id
				) FILTER (WHERE t.id IS NOT NULL),
				 '[]'::json
			) AS tasks
		FROM users u
		LEFT JOIN tasks t ON u.id = t.user_id
		GROUP BY u.id
		ORDER BY u.id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.UserWithTasks, 0)

	for rows.Next() {
		var user model.UserWithTasks
		var tasksJson []byte

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&tasksJson,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(tasksJson, &user.Tasks);
		err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err:= rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}