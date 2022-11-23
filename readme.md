# DynamicSql

__ An SQL driver for golang which can be sued for databasses that frquently changes thier data-source propertis like password for security purposes. This driver will handle such casses without interrupting your connections to DB __


## What Databases it Support?

Presently, it supports Postgres and MySQL databases. However, future support to various different databases is projected.

## How does it work?

This driver basically wraps around your actual driver to be used (i.e Postgres/MySQL driver). Whenever DynamicSql driver notices that the DB password is changed, it aligns any new connection to this DB with new set of configurations.
fsnotify is used to watch over the file which is expected to have latest DB dsn stored. 

traditionally, opening a connection to DB is like below
``` sql.Register("postgres", pq.Driver{})```

initially the postgres driver has to be registerd with SQL package. 

then,

```sql.Open("postgres","postgres://postgres:postgres@host.com")```

here, first argument to ```sql.Open``` is the driver name (which will help sql package to identify which driver to be used. We have already register a driver with this name just above) and second is the actual DSN.

Now, to use DynamicSQl, check the following code

``` 
package main

import (
	"database/sql"
	"log"

	"github.com/abhishekbodhekar/dynamicsql"
	"github.com/lib/pq"
)

func main() {

	dynamicsql.RegisterDriver(pq.Driver{})

	db, err := sql.Open("dynamicsql", "dynamicsql://dummyHost/Users/abhishek/Work/Test/test/abc.txt")
	if err == nil {
		if err = db.Ping(); err == nil {
			log.Println("Ping successful")
		}
	}
}
```

Here, first you need to register your actual driver (i.e. ostgres/MySQl) with dynamicSQl.
After this, while db.Open(), first argument must be "dynamicsql" and second must be in a correct format.
This format is an URI. The scehme should be "dynamicsql". The host can be anything. Then the path shouldd be always the actual path  of file containing latest DSN of the DB. 


