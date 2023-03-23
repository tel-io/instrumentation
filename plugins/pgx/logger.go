package pgx

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/tel-io/tel/v2"
)

type methodLogger interface {
	Callback
}

type methodLoggerImpl struct {
	logger tel.Logger

	dumpSQL bool
}

func (m *methodLoggerImpl) Query(ctx context.Context, start pgx.TraceQueryStartData) (context.Context, QueryFn) {
	startTime := time.Now()

	return ctx, func(conn *pgx.Conn, end pgx.TraceQueryEndData) {
		interval := time.Since(startTime)

		if end.Err != nil {
			m.logger.Error("SQL Query", tel.Error(end.Err), tel.String("sql", start.SQL),
				tel.Any("args", logQueryArgs(start.Args)), tel.Duration("duration", interval),
			)

			return
		}

		if m.dumpSQL {
			m.logger.Debug("SQL Query", tel.String("sql", start.SQL), tel.Any("args",
				logQueryArgs(start.Args)), tel.Duration("duration", interval),
				tel.String("commandTag", end.CommandTag.String()), tel.Uint32("pid", getPID(conn)),
			)
		}
	}
}

func (m *methodLoggerImpl) Batch(ctx context.Context, _ pgx.TraceBatchStartData) (context.Context, BatchQueryFn, BatchEndFn) {
	startTime := time.Now()

	return ctx,
		func(conn *pgx.Conn, data pgx.TraceBatchQueryData) {
			interval := time.Since(startTime)

			if data.Err != nil {
				m.logger.Error("SQL BatchQuery", tel.Error(data.Err), tel.String("sql", data.SQL),
					tel.Any("args", logQueryArgs(data.Args)), tel.Duration("duration", interval),
				)

				return
			}
			if m.dumpSQL {
				m.logger.Debug("SQL BatchQuery", tel.String("commandTag", data.CommandTag.String()), tel.String("sql", data.SQL),
					tel.Any("args", logQueryArgs(data.Args)), tel.Duration("duration", interval),
				)
			}
		},
		func(conn *pgx.Conn, data pgx.TraceBatchEndData) {
			interval := time.Since(startTime)

			if data.Err != nil {
				m.logger.Error("SQL BatchClose", tel.Error(data.Err), tel.Duration("duration", interval))
				return
			}

			if m.dumpSQL {
				m.logger.Debug("SQL BatchClose", tel.Duration("duration", interval))
			}
		}
}

func (m *methodLoggerImpl) Copy(ctx context.Context, start pgx.TraceCopyFromStartData) (context.Context, CopyFn) {
	startTime := time.Now()

	return ctx, func(conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
		interval := time.Since(startTime)

		if data.Err != nil {
			m.logger.Error("SQL CopyFrom", tel.Error(data.Err), tel.Any("tableName", start.TableName),
				tel.Any("columnNames", start.ColumnNames), tel.Duration("duration", interval),
			)

			return
		}

		if m.dumpSQL {
			m.logger.Debug("SQL CopyFrom", tel.Any("tableName", start.TableName),
				tel.Any("columnNames", start.ColumnNames), tel.Duration("duration", interval),
				tel.Int64("rowCount", data.CommandTag.RowsAffected()),
			)
		}
	}
}

func (m *methodLoggerImpl) Connect(ctx context.Context, start pgx.TraceConnectStartData) (context.Context, ConnectFn) {
	startTime := time.Now()

	return ctx, func(data pgx.TraceConnectEndData) {
		interval := time.Since(startTime)

		if data.Err != nil {
			m.logger.Error("SQL Connect", tel.Error(data.Err), tel.Duration("duration", interval),
				tel.String("host", start.ConnConfig.Host),
				tel.Uint16("host", start.ConnConfig.Port),
				tel.String("database", start.ConnConfig.Database),
			)

			return
		}

		if m.dumpSQL {
			m.logger.Debug("SQL Connect", tel.Duration("duration", interval),
				tel.String("host", start.ConnConfig.Host),
				tel.Uint16("host", start.ConnConfig.Port),
				tel.String("database", start.ConnConfig.Database),
			)
		}
	}
}

func (m *methodLoggerImpl) Prepare(ctx context.Context, start pgx.TracePrepareStartData) (context.Context, PrepareFn) {
	startTime := time.Now()

	return ctx, func(conn *pgx.Conn, data pgx.TracePrepareEndData) {
		interval := time.Since(startTime)

		if data.Err != nil {
			m.logger.Error("SQL Prepare", tel.Error(data.Err), tel.String("sql", start.SQL),
				tel.String("name", start.Name), tel.Duration("duration", interval),
			)

			return
		}

		if m.dumpSQL {
			m.logger.Debug("SQL Prepare", tel.Error(data.Err), tel.String("sql", start.SQL),
				tel.String("name", start.Name), tel.Duration("duration", interval),
				tel.Bool("alreadyPrepared", data.AlreadyPrepared),
			)
		}
	}
}

func logQueryArgs(args []any) []any {
	logArgs := make([]any, 0, len(args))

	for _, a := range args {
		switch v := a.(type) {
		case []byte:
			if len(v) < 64 {
				a = hex.EncodeToString(v)
			} else {
				a = fmt.Sprintf("%x (truncated %d bytes)", v[:64], len(v)-64)
			}
		case string:
			if len(v) > 64 {
				var l int = 0
				for w := 0; l < 64; l += w {
					_, w = utf8.DecodeRuneInString(v[l:])
				}
				if len(v) > l {
					a = fmt.Sprintf("%s (truncated %d bytes)", v[:l], len(v)-l)
				}
			}
		}
		logArgs = append(logArgs, a)
	}

	return logArgs
}
