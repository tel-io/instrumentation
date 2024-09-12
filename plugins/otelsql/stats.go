package otelsql

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// defaultMinimumReadDBStatsInterval is the default minimum interval between calls to db.Stats().
const defaultMinimumReadDBStatsInterval = time.Second

// RecordStats records database statistics for provided sql.DB at the provided interval.
func RecordStats(db *sql.DB, opts ...StatsOption) error {
	o := statsOptions{
		meterProvider:              otel.GetMeterProvider(),
		minimumReadDBStatsInterval: defaultMinimumReadDBStatsInterval,
	}

	for _, opt := range opts {
		opt.applyStatsOptions(&o)
	}

	meter := o.meterProvider.Meter(instrumentationName)

	return recordStats(meter, db, o.minimumReadDBStatsInterval, o.defaultAttributes...)
}

// nolint: funlen
func recordStats(
	meter metric.Meter,
	db *sql.DB,
	minimumReadDBStatsInterval time.Duration,
	attrs ...attribute.KeyValue,
) error {
	var (
		err error

		openConnections   metric.Int64ObservableGauge
		idleConnections   metric.Int64ObservableGauge
		activeConnections metric.Int64ObservableGauge
		waitCount         metric.Int64ObservableGauge
		waitDuration      metric.Float64ObservableGauge
		idleClosed        metric.Int64ObservableGauge
		lifetimeClosed    metric.Int64ObservableGauge

		dbStats     sql.DBStats
		lastDBStats time.Time

		// lock prevents a race between batch observer and instrument registration.
		lock sync.Mutex
	)

	lock.Lock()
	defer lock.Unlock()

	if openConnections, err = meter.Int64ObservableGauge(
		dbSQLConnectionsOpen,
		metric.WithUnit("1"),
		metric.WithDescription("Count of open connections in the pool"),
	); err != nil {
		return err
	}

	if idleConnections, err = meter.Int64ObservableGauge(
		dbSQLConnectionsIdle,
		metric.WithUnit("1"),
		metric.WithDescription("Count of idle connections in the pool"),
	); err != nil {
		return err
	}

	if activeConnections, err = meter.Int64ObservableGauge(
		dbSQLConnectionsActive,
		metric.WithUnit("1"),
		metric.WithDescription("Count of active connections in the pool"),
	); err != nil {
		return err
	}

	if waitCount, err = meter.Int64ObservableGauge(
		dbSQLConnectionsWaitCount,
		metric.WithUnit("1"),
		metric.WithDescription("The total number of connections waited for"),
	); err != nil {
		return err
	}

	if waitDuration, err = meter.Float64ObservableGauge(
		dbSQLConnectionsWaitDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("The total time blocked waiting for a new connection"),
	); err != nil {
		return err
	}

	if idleClosed, err = meter.Int64ObservableGauge(
		dbSQLConnectionsIdleClosed,
		metric.WithUnit("1"),
		metric.WithDescription("The total number of connections closed due to SetMaxIdleConns"),
	); err != nil {
		return err
	}

	if lifetimeClosed, err = meter.Int64ObservableGauge(
		dbSQLConnectionsLifetimeClosed,
		metric.WithUnit("1"),
		metric.WithDescription("The total number of connections closed due to SetConnMaxLifetime"),
	); err != nil {
		return err
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		lock.Lock()
		defer lock.Unlock()

		now := time.Now()
		if now.Sub(lastDBStats) >= minimumReadDBStatsInterval {
			dbStats = db.Stats()
			lastDBStats = now
		}

		o.ObserveInt64(openConnections, int64(dbStats.OpenConnections),
			metric.WithAttributes(attrs...),
		)

		o.ObserveInt64(idleConnections, int64(dbStats.Idle),
			metric.WithAttributes(attrs...))

		o.ObserveInt64(activeConnections, int64(dbStats.InUse),
			metric.WithAttributes(attrs...))

		o.ObserveInt64(waitCount, dbStats.WaitCount,
			metric.WithAttributes(attrs...))

		o.ObserveFloat64(waitDuration, float64(dbStats.WaitDuration.Nanoseconds())/1e6,
			metric.WithAttributes(attrs...))

		o.ObserveInt64(idleClosed, dbStats.MaxIdleClosed,
			metric.WithAttributes(attrs...))

		o.ObserveInt64(lifetimeClosed, dbStats.MaxLifetimeClosed,
			metric.WithAttributes(attrs...))

		return nil
	}, []metric.Observable{
		openConnections,
		idleConnections,
		activeConnections,
		waitCount,
		waitDuration,
		idleClosed,
		lifetimeClosed,
	}...)

	return err
}
