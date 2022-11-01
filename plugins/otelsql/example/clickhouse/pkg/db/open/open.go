package open

import (
	"database/sql"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pkg/errors"
	otelsql "github.com/tel-io/instrumentation/plugins/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var clickhouse = semconv.DBSystemKey.String("clockhouse")

func OpenDB(dsn string) (*sql.DB, error) {
	// Register the otelsql wrapper for the provided postgres driver.
	driverName, err := otelsql.Register("clickhouse",
		otelsql.AllowRoot(),
		otelsql.TraceQueryWithoutArgs(),
		otelsql.TraceRowsClose(),
		otelsql.TraceRowsAffected(),
		otelsql.WithDatabaseName("my_database"), // Optional.
		otelsql.WithSystem(clickhouse),          // Optional.
	)
	if err != nil {
		return nil, err
	}

	// Connect to a Clickhouse database using the postgres driver wrapper.
	// create connection with sql.OpenDB
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, errors.WithMessagef(err, "open")
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	if err = db.Ping(); err != nil {
		return nil, errors.WithMessagef(err, "ping")
	}

	return db, nil
}
