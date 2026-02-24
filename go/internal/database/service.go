package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/kagent-dev/kagent/go/internal/metrics"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Model interface {
	TableName() string
}

type Clause struct {
	Key   string
	Value any
}

func list[T Model](db *gorm.DB, clauses ...Clause) ([]T, error) {
	var models []T
	var t T
	table := t.TableName()
	start := time.Now()

	query := db

	for _, clause := range clauses {
		query = query.Where(fmt.Sprintf("%s = ?", clause.Key), clause.Value)
	}

	err := query.Order("created_at ASC").Find(&models).Error
	metrics.DatabaseOperationsTotal.WithLabelValues("list", table).Inc()
	metrics.DatabaseOperationDuration.WithLabelValues("list", table).Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues("list", table).Inc()
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	return models, nil
}

func get[T Model](db *gorm.DB, clauses ...Clause) (*T, error) {
	var model T
	table := model.TableName()
	start := time.Now()

	query := db

	for _, clause := range clauses {
		query = query.Where(fmt.Sprintf("%s = ?", clause.Key), clause.Value)
	}

	err := query.First(&model).Error
	metrics.DatabaseOperationsTotal.WithLabelValues("get", table).Inc()
	metrics.DatabaseOperationDuration.WithLabelValues("get", table).Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues("get", table).Inc()
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	return &model, nil
}

// save performs an upsert operation (INSERT ON CONFLICT DO UPDATE)
// args:
// - db: the database connection
// - model: the model to save
func save[T Model](db *gorm.DB, model *T) error {
	table := (*model).TableName()
	start := time.Now()

	if err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(model).Error; err != nil {
		metrics.DatabaseOperationsTotal.WithLabelValues("save", table).Inc()
		metrics.DatabaseOperationDuration.WithLabelValues("save", table).Observe(time.Since(start).Seconds())
		metrics.DatabaseErrors.WithLabelValues("save", table).Inc()
		return fmt.Errorf("failed to upsert model: %w", err)
	}

	metrics.DatabaseOperationsTotal.WithLabelValues("save", table).Inc()
	metrics.DatabaseOperationDuration.WithLabelValues("save", table).Observe(time.Since(start).Seconds())
	return nil
}

func delete[T Model](db *gorm.DB, clauses ...Clause) error {
	t := new(T)
	table := (*t).TableName()
	start := time.Now()

	query := db

	for _, clause := range clauses {
		query = query.Where(fmt.Sprintf("%s = ?", clause.Key), clause.Value)
	}

	result := query.Delete(t)
	metrics.DatabaseOperationsTotal.WithLabelValues("delete", table).Inc()
	metrics.DatabaseOperationDuration.WithLabelValues("delete", table).Observe(time.Since(start).Seconds())
	if result.Error != nil {
		metrics.DatabaseErrors.WithLabelValues("delete", table).Inc()
		return fmt.Errorf("failed to delete model: %w", result.Error)
	}
	return nil
}

// BuildWhereClause is deprecated, use individual Where clauses instead
func BuildWhereClause(clauses ...Clause) string {
	var clausesStr strings.Builder
	for idx, clause := range clauses {
		if idx > 0 {
			clausesStr.WriteString(" AND ")
		}
		clausesStr.WriteString(fmt.Sprintf("%s = %v", clause.Key, clause.Value))
	}
	return clausesStr.String()
}
