package good

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go"
)

type PgStore struct {
	Pool   *pgxpool.Pool
	logger *nats.Conn
}

func NewStore(Pool *pgxpool.Pool, logger *nats.Conn) Store {
	return PgStore{
		Pool:   Pool,
		logger: logger,
	}
}

func (s PgStore) logToClickHouse(good Good) {
	payload, err := json.Marshal(struct {
		Id          int64  `json:"Id"`
		ProjectId   int64  `json:"ProjectId"`
		Name        string `json:"Name"`
		Description string `json:"Description"`
		Priority    int64  `json:"Priority"`
		Removed     bool   `json:"Removed"`
		CreatedAt   string `json:"EventTime"`
	}{
		Id:          good.Id,
		ProjectId:   good.ProjectId,
		Name:        good.Name,
		Description: good.Description,
		Priority:    good.Priority,
		Removed:     good.Removed,
		CreatedAt:   good.CreatedAt.Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		log.Printf("failed to marshal payload for logging: %v", err)
	} else {
		err = s.logger.Publish("logs", payload)
		if err != nil {
			log.Printf("failed to log to nats: %v", err)
		}
	}
}

func (s PgStore) Count(ctx context.Context) int64 {
	var count int64
	row := s.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM goods`)
	err := row.Scan(&count)
	if err != nil {
		return 0
	}

	return count
}

func (s PgStore) RemovedCount(ctx context.Context) int64 {
	var count int64
	row := s.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM goods WHERE removed = true`)
	err := row.Scan(&count)
	if err != nil {
		return 0
	}

	return count
}

func (s PgStore) ListGoods(ctx context.Context, params ListGoodsParams) ([]Good, error) {
	rows, err := s.Pool.Query(
		ctx,
		`SELECT 
		id,
		project_id,
		name,
		description,
		priority,
		removed,
		created_at
	FROM
		goods
	ORDER BY id
	LIMIT $1 OFFSET $2`,
		params.Limit,
		params.Offset,
	)

	if err != nil {
		return nil, err
	}

	var goods []Good
	for rows.Next() {
		var good Good
		if err := rows.Scan(
			&good.Id,
			&good.ProjectId,
			&good.Name,
			&good.Description,
			&good.Priority,
			&good.Removed,
			&good.CreatedAt,
		); err != nil {
			return nil, err
		}

		goods = append(goods, good)
	}
	rows.Close()

	return goods, nil
}

func (s PgStore) CreateGood(ctx context.Context, good Good) (Good, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Good{}, fmt.Errorf("failed to begin tx: %w", err)
	}

	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(1)")
	if err != nil {
		return Good{}, fmt.Errorf("failed to acquire advisory tx lock: %w", err)
	}

	var maxPriority int64
	err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(priority), 0) FROM goods`).Scan(&maxPriority)
	maxPriority++

	row := tx.QueryRow(
		ctx,
		`INSERT INTO goods (
        id,
		project_id,
    	name,
        description,
    	priority,
        removed,
    	created_at
	)
	VALUES
		($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, project_id, name, description, priority, removed, created_at`,
		good.Id,
		good.ProjectId,
		good.Name,
		good.Description,
		maxPriority,
		good.Removed,
		good.CreatedAt,
	)

	var createdGood Good
	err = row.Scan(
		&createdGood.Id,
		&createdGood.ProjectId,
		&createdGood.Name,
		&createdGood.Description,
		&createdGood.Priority,
		&createdGood.Removed,
		&createdGood.CreatedAt,
	)
	if err != nil {
		return Good{}, fmt.Errorf("failed to scan rows: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		err = tx.Rollback(ctx)
		if err != nil {
			return Good{}, fmt.Errorf("failed to rollback tx: %w", err)
		}
		return Good{}, fmt.Errorf("failed to commit tx: %w, rollbacked successfully", err)
	}

	s.logToClickHouse(good)

	return createdGood, nil
}

func (s PgStore) DeleteGood(ctx context.Context, id, projectId int64) (Good, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Good{}, fmt.Errorf("failed to begin tx: %w", err)
	}

	row := tx.QueryRow(
		ctx,
		`SELECT 
		id,
		project_id,
		name,
		description,
		priority,
		removed,
		created_at
	FROM
		goods
	WHERE id = $1 AND project_id = $2
	FOR UPDATE`,
		id,
		projectId,
	)

	var good Good
	err = row.Scan(
		&good.Id,
		&good.ProjectId,
		&good.Name,
		&good.Description,
		&good.Priority,
		&good.Removed,
		&good.CreatedAt,
	)
	if err != nil {
		return Good{}, err
	}

	row = tx.QueryRow(
		ctx,
		`UPDATE goods 
	SET removed = true 
	WHERE id = $1 AND project_id = $2
	RETURNING id, project_id, name, description, priority, removed, created_at`,
		id,
		projectId,
	)

	if err != nil {
		return Good{}, fmt.Errorf("failed to update row: %w", err)
	}

	err = row.Scan(
		&good.Id,
		&good.ProjectId,
		&good.Name,
		&good.Description,
		&good.Priority,
		&good.Removed,
		&good.CreatedAt,
	)

	err = tx.Commit(ctx)
	if err != nil {
		err = tx.Rollback(ctx)
		if err != nil {
			return Good{}, fmt.Errorf("failed to rollback tx: %w", err)
		}
		return Good{}, fmt.Errorf("failed to commit tx: %w, rollbacked successfully", err)
	}

	s.logToClickHouse(good)

	return good, nil
}

func (s PgStore) UpdateGood(ctx context.Context, id, projectId int64, params UpdateGoodParams) (Good, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Good{}, fmt.Errorf("failed to begin tx: %w", err)
	}

	row := tx.QueryRow(
		ctx,
		`SELECT 
		id,
		project_id,
		name,
		description,
		priority,
		removed,
		created_at
	FROM
		goods
	WHERE id = $1 AND project_id = $2
	FOR UPDATE`,
		id,
		projectId,
	)

	var good Good
	err = row.Scan(
		&good.Id,
		&good.ProjectId,
		&good.Name,
		&good.Description,
		&good.Priority,
		&good.Removed,
		&good.CreatedAt,
	)
	if err != nil {
		return Good{}, err
	}

	description := good.Description
	if params.Description != "" {
		description = params.Description
	}

	row = tx.QueryRow(
		ctx,
		`UPDATE goods 
	SET 
	    project_id = $1,
	    name = $2,
		description = $3,
	    priority = $4,
	    removed = $5
	WHERE id = $6 and project_id = $7
	RETURNING id, project_id, name, description, priority, removed, created_at`,
		params.ProjectId,
		params.Name,
		description,
		params.Priority,
		params.Removed,
		id,
		projectId,
	)
	if err != nil {
		return Good{}, fmt.Errorf("failed to update row: %w", err)
	}

	err = row.Scan(
		&good.Id,
		&good.ProjectId,
		&good.Name,
		&good.Description,
		&good.Priority,
		&good.Removed,
		&good.CreatedAt,
	)

	err = tx.Commit(ctx)
	if err != nil {
		err = tx.Rollback(ctx)
		if err != nil {
			return Good{}, fmt.Errorf("failed to rollback tx: %w", err)
		}
		return Good{}, fmt.Errorf("failed to commit tx: %w, rollbacked successfully", err)
	}

	s.logToClickHouse(good)

	return good, nil
}

func (s PgStore) ReprioritizeGood(ctx context.Context, id, projectId int64, params ReprioritizeGoodParams) (Good, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Good{}, fmt.Errorf("failed to begin tx: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
		WITH updated_row AS (
			UPDATE goods
			SET priority = $1
			WHERE id = $2
			RETURNING id
		), subsequent_rows AS (
			SELECT id, priority
			FROM goods
			WHERE id > (SELECT id FROM updated_row)
			ORDER BY id
		), new_priorities AS (
			SELECT id, $1 + ROW_NUMBER() OVER (ORDER BY id) AS new_priority
			FROM subsequent_rows
		)
		UPDATE goods
		SET priority = new_priorities.new_priority
		FROM new_priorities
		WHERE goods.id = new_priorities.id`,
		params.NewPriority,
		id,
	)
	if err != nil {
		return Good{}, fmt.Errorf("failed to update row: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		err = tx.Rollback(ctx)
		if err != nil {
			return Good{}, fmt.Errorf("failed to rollback tx: %w", err)
		}
		return Good{}, fmt.Errorf("failed to commit tx: %w, rollbacked successfully", err)
	}

	row := s.Pool.QueryRow(
		ctx,
		`
    	SELECT id, project_id, name, description, priority, removed, created_at
    	FROM goods
    	WHERE id = $1 AND project_id = $2`,
		id,
		projectId,
	)

	var good Good
	err = row.Scan(
		&good.Id,
		&good.ProjectId,
		&good.Name,
		&good.Description,
		&good.Priority,
		&good.Removed,
		&good.CreatedAt,
	)

	if err != nil {
		return Good{}, err
	}

	s.logToClickHouse(good)

	return good, nil
}

func (s PgStore) GetReprioritizedGoods(ctx context.Context, id int64) ([]ReprioritizedGood, error) {
	rows, err := s.Pool.Query(
		ctx,
		`SELECT 
		id,
		priority
		FROM
			goods
		WHERE id >= $1
		ORDER BY id`,
		id,
	)

	if err != nil {
		return nil, err
	}

	var goods []ReprioritizedGood
	for rows.Next() {
		var good ReprioritizedGood
		if err := rows.Scan(
			&good.Id,
			&good.Priority,
		); err != nil {
			return nil, err
		}

		goods = append(goods, good)
	}
	rows.Close()

	return goods, nil
}
