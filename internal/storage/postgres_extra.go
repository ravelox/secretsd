package storage

import "database/sql"

func (p *Postgres) UnderlyingDB() *sql.DB { return p.db }
