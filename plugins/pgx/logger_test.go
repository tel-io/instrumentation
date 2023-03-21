package pgx

import (
	"bufio"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/stretchr/testify/assert"
	"github.com/tel-io/tel/v2"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name  string
		level tracelog.LogLevel
		data  map[string]interface{}
		check []string
	}{
		{
			"LogLevelTrace",
			tracelog.LogLevelTrace,
			map[string]interface{}{"X": true},
			[]string{"PGX_LOG_LEVEL", "trace"},
		},
		{
			"LogLevelDebug",
			tracelog.LogLevelDebug,
			map[string]interface{}{"PGX_LOG_LEVEL": true},
			[]string{"PGX_LOG_LEVEL", "true"},
		},
		{
			"LogLevelInfo",
			tracelog.LogLevelInfo,
			map[string]interface{}{"PGX_LOG_LEVEL": true},
			[]string{"PGX_LOG_LEVEL", "true"},
		},
		{
			"LogLevelWarn",
			tracelog.LogLevelWarn,
			map[string]interface{}{"PGX_LOG_LEVEL": true},
			[]string{"PGX_LOG_LEVEL", "true"},
		},
		{
			"LogLevelError",
			tracelog.LogLevelError,
			map[string]interface{}{"PGX_LOG_LEVEL": true},
			[]string{"PGX_LOG_LEVEL", "true"},
		},
		{
			"check sql and args fields",
			tracelog.LogLevelInfo,
			map[string]interface{}{fSql: "insert * from table where user = $1", fArgs: []interface{}{100500}},
			[]string{"insert * from table where user = 100500"},
		},
		{
			"check sql no args",
			tracelog.LogLevelInfo,
			map[string]interface{}{fSql: "insert * from table where user = $1"},
			[]string{"insert * from table where user = $"},
		},
		{
			"multi-line",
			tracelog.LogLevelInfo,
			map[string]interface{}{fSql: `UPDATE tx SET revert = true WHERE
						created_at < current_timestamp  AND  created_at > current_timestamp - interval '3' month AND
						id = $1 AND "accountId" = $2 `},
			[]string{"UPDATE tx SET revert = true WHERE created_at < current_timestamp AND created_at > current_timestamp - interval"},
		},
	}

	tt := tel.NewNull()

	buf := tel.SetLogOutput(&tt)
	ctx := tt.Ctx()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			NewLogger().Log(ctx, test.level, "test", test.data)

			line, _, err := bufio.NewReader(buf).ReadLine()
			assert.NoError(t, err)
			fmt.Println(string(line))

			for _, val := range test.check {
				assert.Contains(t, string(line), val)
			}
		})
	}
}
