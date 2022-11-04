package sdk

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// Compile-time proof of interface implementation.
var _ Schemas = (*schemas)(nil)

// Schemas describes all the schemas related methods that the
// Snowflake API supports.
type Schemas interface {
	// List all the schemas by pattern.
	List(ctx context.Context, options SchemaListOptions) ([]*Schema, error)
	// Create a new schema with the given options.
	Create(ctx context.Context, options SchemaCreateOptions) (*Schema, error)
	// Read a schema by its name.
	Read(ctx context.Context, schema string) (*Schema, error)
	// Update attributes of an existing schema.
	Update(ctx context.Context, schema string, options SchemaUpdateOptions) (*Schema, error)
	// Delete a schema by its name.
	Delete(ctx context.Context, schema string) error
}

// schemas implements Schemas
type schemas struct {
	client *Client
}

// Schema represents a Snowflake schema.
type Schema struct {
	Name          string
	IsDefault     string
	IsCurrent     string
	DatabaseName  string
	Owner         string
	Comment       string
	Options       string
	RetentionTime string
	CreatedOn     time.Time
}

type schemaEntity struct {
	Name          sql.NullString `db:"name"`
	IsDefault     sql.NullString `db:"is_default"`
	IsCurrent     sql.NullString `db:"is_current"`
	DatabaseName  sql.NullString `db:"database_name"`
	Owner         sql.NullString `db:"owner"`
	Comment       sql.NullString `db:"comment"`
	Options       sql.NullString `db:"options"`
	RetentionTime sql.NullString `db:"retention_time"`
	CreatedOn     sql.NullTime   `db:"created_on"`
}

func (s *schemaEntity) toSchema() *Schema {
	return &Schema{
		Name:          s.Name.String,
		IsDefault:     s.IsDefault.String,
		IsCurrent:     s.IsCurrent.String,
		DatabaseName:  s.DatabaseName.String,
		Owner:         s.Owner.String,
		Comment:       s.Comment.String,
		Options:       s.Options.String,
		RetentionTime: s.RetentionTime.String,
		CreatedOn:     s.CreatedOn.Time,
	}
}

// SchemaListOptions represents the options for listing schemas.
type SchemaListOptions struct {
	Pattern string
}

func (o SchemaListOptions) validate() error {
	if o.Pattern == "" {
		return errors.New("pattern must not be empty")
	}
	return nil
}

type SchemaProperties struct {
	// Optional: Specifies a comment for the schema.
	Comment *string
}

// SchemaCreateOptions represents the options for creating a schema.
type SchemaCreateOptions struct {
	*SchemaProperties

	// Required: Specifies the identifier for the schema; must be unique for the database in which the schema is created.
	Name string

	DatabaseName string
}

func (o SchemaCreateOptions) validate() error {
	if o.Name == "" {
		return errors.New("schema name must not be empty")
	}
	if o.DatabaseName == "" {
		return errors.New("database name must not be empty")
	}
	return nil
}

// SchemaUpdateOptions represents the options for updating a schema.
type SchemaUpdateOptions struct {
	*SchemaProperties
}

// List all the schemas by pattern.
func (s *schemas) List(ctx context.Context, options SchemaListOptions) ([]*Schema, error) {
	if err := options.validate(); err != nil {
		return nil, fmt.Errorf("validate list options: %w", err)
	}

	sql := fmt.Sprintf(`SHOW SCHEMAS LIKE '%s'`, options.Pattern)
	rows, err := s.client.query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("do query: %w", err)
	}
	defer rows.Close()

	entities := []*Schema{}
	for rows.Next() {
		var entity schemaEntity
		if err := rows.StructScan(&entity); err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		entities = append(entities, entity.toSchema())
	}
	return entities, nil
}

// Create a new schema with the given options.
func (s *schemas) Create(ctx context.Context, options SchemaCreateOptions) (*Schema, error) {
	if err := options.validate(); err != nil {
		return nil, fmt.Errorf("validate create options: %w", err)
	}
	sql := fmt.Sprintf("USE DATABASE %s", options.DatabaseName)
	if _, err := s.client.exec(ctx, sql); err != nil {
		return nil, fmt.Errorf("db exec: %w", err)
	}

	sql = fmt.Sprintf("CREATE SCHEMA %s", options.Name)
	if options.SchemaProperties != nil {
		sql = sql + s.formatSchemaProperties(options.SchemaProperties)
	}
	if _, err := s.client.exec(ctx, sql); err != nil {
		return nil, fmt.Errorf("db exec: %w", err)
	}
	return s.Read(ctx, options.Name)
}

// Read a schema by its name.
func (s *schemas) Read(ctx context.Context, schema string) (*Schema, error) {
	sql := fmt.Sprintf(`SHOW SCHEMAS LIKE '%s'`, schema)
	rows, err := s.client.query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("do query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}
	var entity schemaEntity
	if err := rows.StructScan(&entity); err != nil {
		return nil, fmt.Errorf("rows scan: %w", err)
	}
	return entity.toSchema(), nil
}

func (*schemas) formatSchemaProperties(properties *SchemaProperties) string {
	var s string
	if properties.Comment != nil {
		s = s + " comment='" + *properties.Comment + "'"
	}
	return s
}

// Update attributes of an existing schema.
func (s *schemas) Update(ctx context.Context, schema string, options SchemaUpdateOptions) (*Schema, error) {
	if schema == "" {
		return nil, errors.New("name must not be empty")
	}
	sql := fmt.Sprintf("ALTER SCHEMA %s SET", schema)
	if options.SchemaProperties != nil {
		sql = sql + s.formatSchemaProperties(options.SchemaProperties)
	}
	if _, err := s.client.exec(ctx, sql); err != nil {
		return nil, fmt.Errorf("db exec: %w", err)
	}
	return s.Read(ctx, schema)
}

// Delete a schema by its name.
func (s *schemas) Delete(ctx context.Context, schema string) error {
	sql := fmt.Sprintf(`DROP SCHEMA %s`, schema)
	if _, err := s.client.exec(ctx, sql); err != nil {
		return fmt.Errorf("db exec: %w", err)
	}
	return nil
}
