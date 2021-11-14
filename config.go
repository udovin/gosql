package gosql

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
)

type SQLiteConfig struct {
	Path string
	Mode string
}

func (c SQLiteConfig) NewDB() (*DB, error) {
	rw, err := c.newDB()
	if err != nil {
		return nil, err
	}
	db := DB{
		DB:      rw,
		RO:      rw,
		Builder: NewBuilder(SQLiteDialect),
	}
	return &db, nil
}

func (c SQLiteConfig) newDB() (*sql.DB, error) {
	params := url.Values{}
	if c.Mode != "" {
		params.Add("mode", c.Mode)
	}
	if c.Mode == "memory" || c.Path == ":memory:" {
		params.Set("cache", "shared")
	}
	return sql.Open("sqlite3", fmt.Sprintf(
		"file:%s?%s", c.Path, params.Encode(),
	))
}

type PostgresConfig struct {
	Hosts    []string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c PostgresConfig) NewDB() (*DB, error) {
	rw, err := c.newDB(true)
	if err != nil {
		return nil, err
	}
	ro, err := c.newDB(false)
	if err != nil {
		_ = rw.Close()
		return nil, err
	}
	db := DB{
		DB:      rw,
		RO:      ro,
		Builder: NewBuilder(PostgresDialect),
	}
	return &db, nil
}

func (c PostgresConfig) buildHostPort(builder *strings.Builder) {
	var hosts strings.Builder
	var ports strings.Builder
	for _, host := range c.Hosts {
		parts := strings.SplitN(host, ":", 2)
		if len(parts) == 0 {
			continue
		}
		if hosts.Len() > 0 {
			hosts.WriteRune(',')
			ports.WriteRune(',')
		}
		hosts.WriteString(parts[0])
		if len(parts) > 1 {
			ports.WriteString(parts[1])
		} else {
			ports.WriteString("5432")
		}
	}
	builder.WriteString("host=")
	builder.WriteString(hosts.String())
	builder.WriteString(" port=")
	builder.WriteString(ports.String())
}

func (c PostgresConfig) newDB(writable bool) (*sql.DB, error) {
	var connStr strings.Builder
	c.buildHostPort(&connStr)
	connStr.WriteString(" user=")
	connStr.WriteString(c.User)
	connStr.WriteString(" password=")
	connStr.WriteString(c.Password)
	connStr.WriteString(" dbname=")
	connStr.WriteString(c.Name)
	connStr.WriteString(" statement_cache_mode=describe")
	if c.SSLMode != "" {
		connStr.WriteString(" sslmode=")
		connStr.WriteString(c.SSLMode)
	}
	connStr.WriteString(" statement_timeout=120000")                   // 2 min.
	connStr.WriteString(" idle_in_transaction_session_timeout=120000") // 2 min.
	if writable {
		connStr.WriteString(" target_session_attrs=read-write")
	}
	return sql.Open("pgx", connStr.String())
}
