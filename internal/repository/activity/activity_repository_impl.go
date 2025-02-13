package activity

import (
	"database/sql"
	"fmt"

	"github.com/vnnyx/golang-todo-api/internal/infrastructure"
	"github.com/vnnyx/golang-todo-api/internal/model/entity"
)

type ActivityRepositoryImpl struct {
	db *sql.DB
}

func NewActivityRepository(db *sql.DB) ActivityRepository {
	return &ActivityRepositoryImpl{db: db}
}

func (repo *ActivityRepositoryImpl) InsertActivity(activity entity.Activity) (*entity.Activity, error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "INSERT INTO activities(title, email) VALUES(?,?)"

	args := []interface{}{
		activity.Title,
		activity.Email,
	}
	result, err := repo.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	a, err := repo.GetActivityByID(id)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (repo *ActivityRepositoryImpl) GetActivityByID(id int64) (activity *entity.Activity, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "SELECT * FROM activities WHERE activity_id=?"
	rows, err := repo.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var a entity.Activity
		err = rows.Scan(&a.ID, &a.Title, &a.Email, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		return &a, nil
	}
	return nil, fmt.Errorf("Activity with ID %v Not Found", id)
}

func (repo *ActivityRepositoryImpl) GetAllActivity() (activities []*entity.Activity, err error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "SELECT * FROM activities"
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a entity.Activity
		err = rows.Scan(&a.ID, &a.Title, &a.Email, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		activities = append(activities, &a)
	}
	return activities, nil
}

func (repo *ActivityRepositoryImpl) UpdateActivity(activity entity.Activity) (*entity.Activity, error) {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "UPDATE activities SET title=? WHERE activity_id=?"
	args := []interface{}{
		activity.Title,
		activity.ID,
	}
	_, err := repo.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	a, err := repo.GetActivityByID(activity.ID)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (repo *ActivityRepositoryImpl) DeleteActivity(id int64) error {
	ctx, cancel := infrastructure.NewMySQLContext()
	defer cancel()

	query := "DELETE FROM activities WHERE activity_id=?"
	_, err := repo.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
