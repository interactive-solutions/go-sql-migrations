Go migrations
=============

A simple library with the primary objective of being injected into projects startup
executes a set of migration files based upon version name and name.


## Example integration
```go
package main

import (
	"github.com/interactive-solutions/go-sql-migrate"
	"github.com/interactive-solutions/go-sql-migrate/driver"
)

func migrate(host, user, password, dbname string) {

	url := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)

	conn, err := sqlx.Connect("driver", url)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	mig := migrations.CreateFromDirectory("./pkg/database/postgres/migrations")
	mig.Up(driver.NewPostgresDriver(conn))
}
```

## Todo list
- [ ] Add a utility method to create a migration file from current time and provide path
- [ ] Migrate down `x` versions
