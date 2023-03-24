//go:build development
// +build development

package pgx

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/tel-io/tel/v2"
)

func TestXXX(t *testing.T) {
	l, closer := tel.New(context.Background(), tel.DefaultDebugConfig())
	defer closer()

	buf := tel.SetLogOutput(&l)

	l.Debug("XXX")

	c, err := pgx.ParseConfig("postgresql://postgres@localhost/db_test?sslmode=disable")
	require.NoError(t, err)

	// possible we don't need sqlwrapper after that
	c.Tracer, err = New(WithLoggerDumpSQL(true), WithTraceRoot(true))
	require.NoError(t, err)

	conector := stdlib.GetConnector(*c)

	db := sql.OpenDB(conector)

	///
	t.Run("tx", func(t *testing.T) {
		span, ctx := l.StartSpan(l.Ctx(), "TX")
		defer span.End()

		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)

		r, err := tx.QueryContext(ctx, "SELECT 2")
		require.NoError(t, err)

		for r.Next() {
			require.NoError(t, r.Err())

			var i int
			err = r.Scan(&i)
			require.NoError(t, err)
			require.Equal(t, 2, i)
		}
		err = tx.Commit() // i expect this would come to main context
		require.NoError(t, err)
	})

	fmt.Println(buf.String())

	time.Sleep(time.Minute)

}
