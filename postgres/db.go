package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/lib/pq"
	"github.com/ronaudinho/dbbeat/config"
)

type Conn struct {
	conf     config.DB
	DB       *sql.DB
	Listener *pq.Listener
	DBName   string
}

func NewConn(conf config.DB) *Conn {
	return &Conn{
		conf: conf,
	}
}

func (c *Conn) Connect() error {
	db, err := sql.Open("postgres", c.conf.URI)
	if err != nil {
		return err
	}
	log.Println(c.conf.URI)
	c.DB = db
	return nil
}

func (c *Conn) Close() error {
	return c.DB.Close()
}

func (c *Conn) Listen() error {
	query := c.DB.QueryRow("SELECT current_database()")

	query.Scan(&c.DBName)

	log.Println("DBName:", c.DBName)

	err := c.prep()
	if err != nil {
		return err
	}

	reportErr := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("failed to start listener: %s", err)
		}
	}

	listener := pq.NewListener(c.conf.URI, c.conf.MinReconn, c.conf.MaxReconn, reportErr)
	for tbl, ops := range c.conf.TableOps {
		for _, op := range ops {
			tblOp := fmt.Sprintf("%s_%s", tbl, op)
			err = listener.Listen(tblOp)
			if err != nil {
				log.Printf("listen error: channel %s: %v\n", tblOp, err)
			}
			log.Printf("listening: channel %s\n", tblOp)
		}
	}
	c.Listener = listener
	return nil
}

// Prep prepares functions and triggers for a given Conn.
func (c *Conn) prep() error {
	err := c.prepNotifyTrigger()
	if err != nil {
		return err
	}

	to := make(map[string][]string)
	if c.conf.WatchAll {
		tables, err := c.listTables()
		if err != nil {
			return err
		}
		for _, t := range tables {
			to[t] = []string{"insert", "update", "delete"}
		}
		c.conf.TableOps = to
	} else if !c.conf.WatchAll && c.conf.TableOps == nil {
		return errors.New("tables and operations to watch not defined.")
	}
	for tbl, ops := range c.conf.TableOps {
		for _, o := range ops {
			err := c.prepTrigger(tbl, o)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Conn) prepNotifyTrigger() error {
	txn, err := c.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(`
	create or replace function notify_trigger() returns trigger as $$
	declare
		chops text := TG_ARGV[0];
	begin
		PERFORM pg_notify(chops, row_to_json(NEW)::text);
		RETURN NULL;
	end;
	$$ language plpgsql;
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	err = stmt.Close()
	if err != nil {
		return err
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (c *Conn) prepTrigger(channel, ops string) error {
	var exists bool
	err := c.DB.QueryRow((fmt.Sprintf("select count(tgname) from pg_trigger where tgname = %s", pq.QuoteLiteral(fmt.Sprintf("%s_%s", channel, ops))))).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	chOps := fmt.Sprintf("%s_%s", channel, ops)
	_, err = c.DB.Exec(fmt.Sprintf("create trigger %s_%s after %s on %s for each row execute procedure notify_trigger(%s)", channel, ops, ops, channel, pq.QuoteLiteral(chOps)))
	if err != nil {
		return err
	}
	return nil
}

func (c *Conn) listTables() ([]string, error) {
	var vals []string
	rows, err := c.DB.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = $1;", "public")
	if err != nil {
		return vals, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		vals = append(vals, name)
	}
	return vals, nil
}
