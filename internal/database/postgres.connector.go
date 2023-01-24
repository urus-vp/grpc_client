package database

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type PostgresConnector struct {
	pool *sqlx.DB
	dsn  string
}

func NewPostgresConnector(dsn string) (*PostgresConnector, error) {
	pool, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresConnector{pool: pool, dsn: dsn}, nil
}

func (ctor *PostgresConnector) Pool() *sqlx.DB {
	return ctor.pool
}

func (ctor *PostgresConnector) ListenForUpdates(tableName string) (chan bool, error) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, ctor.dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot open connection for listener: %w", err)
	}

	c := make(chan bool)

	_, err = conn.Exec(ctx, fmt.Sprintf(`listen "%s_updates"`, tableName))
	if err != nil {
		return nil, fmt.Errorf("cannot run listener: %w", err)
	}

	ctor.registerTrigger(ctx, tableName)

	go func() {
		for {
			n, err := conn.WaitForNotification(ctx)
			if err != nil {
				log.Errorln("cannot subscribe to notifications:", err)
				close(c)
				return
			}

			log.Debugln("PID:", n.PID, "Channel:", n.Channel, "Payload:", n.Payload)
			c <- true
		}
	}()

	return c, nil
}

func (ctor *PostgresConnector) registerTrigger(ctx context.Context, tableName string) {
	ctor.pool.MustExec(`
		CREATE OR REPLACE FUNCTION do_notify()
		RETURNS trigger AS $$
		BEGIN
			PERFORM pg_notify(TG_ARGV[0], TG_ARGV[0]);
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql;
	`)

	ctor.pool.Exec(fmt.Sprintf(`
		CREATE TRIGGER "notify_%s_update"
		AFTER INSERT OR UPDATE ON "%s"
		FOR EACH STATEMENT EXECUTE PROCEDURE do_notify('%s_updates');
	`, tableName, tableName, tableName))
}
