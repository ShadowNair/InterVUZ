package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertMany(ctx context.Context, items []domain.StructureUnit) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const upsertStructureUnit = `
		INSERT INTO structure_units (
			external_uuid,
			parent_external_uuid,
			code,
			name,
			node_type,
			course,
			semester,
			raw_payload,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (external_uuid) DO UPDATE SET
			parent_external_uuid = EXCLUDED.parent_external_uuid,
			code                 = EXCLUDED.code,
			name                 = EXCLUDED.name,
			node_type            = EXCLUDED.node_type,
			course               = EXCLUDED.course,
			semester             = EXCLUDED.semester,
			raw_payload          = EXCLUDED.raw_payload,
			updated_at           = NOW()
	`

	const upsertAcademicGroup = `
		INSERT INTO academic_groups (
			external_uuid,
			code,
			name,
			faculty_name,
			department_name,
			course,
			semester,
			faculty_external_uuid,
			faculty_code,
			department_external_uuid,
			department_code,
			course_node_external_uuid,
			course_node_name,
			raw_payload,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		ON CONFLICT (external_uuid) DO UPDATE SET
			code                    = EXCLUDED.code,
			name                    = EXCLUDED.name,
			faculty_name            = EXCLUDED.faculty_name,
			department_name         = EXCLUDED.department_name,
			course                  = EXCLUDED.course,
			semester                = EXCLUDED.semester,
			faculty_external_uuid   = EXCLUDED.faculty_external_uuid,
			faculty_code            = EXCLUDED.faculty_code,
			department_external_uuid= EXCLUDED.department_external_uuid,
			department_code         = EXCLUDED.department_code,
			course_node_external_uuid = EXCLUDED.course_node_external_uuid,
			course_node_name        = EXCLUDED.course_node_name,
			raw_payload             = EXCLUDED.raw_payload,
			updated_at              = NOW()
	`

	unitStmt, err := tx.PrepareContext(ctx, upsertStructureUnit)
	if err != nil {
		return fmt.Errorf("prepare structure_units statement: %w", err)
	}
	defer unitStmt.Close()

	groupStmt, err := tx.PrepareContext(ctx, upsertAcademicGroup)
	if err != nil {
		return fmt.Errorf("prepare academic_groups statement: %w", err)
	}
	defer groupStmt.Close()

	for _, item := range items {
		_, err = unitStmt.ExecContext(
			ctx,
			item.ExternalUUID,
			nullableString(item.ParentExternalID),
			item.Code,
			item.Name,
			item.NodeType,
			nullableInt(item.Course),
			nullableInt(item.Semester),
			item.RawPayload,
		)
		if err != nil {
			return fmt.Errorf("upsert structure unit %s: %w", item.ExternalUUID, err)
		}

		if item.NodeType != "group" {
			continue
		}

		_, err = groupStmt.ExecContext(
			ctx,
			item.ExternalUUID,
			item.Code,
			item.Name,
			nullableString(item.FacultyName),
			nullableString(item.DepartmentName),
			nullableInt(item.Course),
			nullableInt(item.Semester),
			nullableString(item.FacultyExternalUUID),
			nullableString(item.FacultyCode),
			nullableString(item.DepartmentExternalUUID),
			nullableString(item.DepartmentCode),
			nullableString(item.CourseNodeExternalUUID),
			nullableString(item.CourseNodeName),
			item.RawPayload,
		)
		if err != nil {
			return fmt.Errorf("upsert academic group %s: %w", item.ExternalUUID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func nullableString(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}