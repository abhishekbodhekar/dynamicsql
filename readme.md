# DynamicSql

__An SQL driver for golang that could be used for databases that frequently change their data-source properties like a password for security purposes. This driver will handle such cases without interrupting your connections to DB__

## What Databases does it Support?

Presently, it supports Postgres and MySQL databases. However, support for different databases is projected in future.

## How does it work?

This driver wraps around your actual driver to be used (i.e Postgres/MySQL driver). Whenever the DynamicSql driver notices that the DB password is changed, it aligns any new connection to this DB with a new set of configurations.
fsnotify is used to watch over the file which is expected to have the latest DB dsn stored. 

traditionally, opening a connection to DB is like below
``` SQL.Register("Postgres", pq.Driver{})```

initially, the Postgres driver has to be registered with the SQL package. 

then,

```SQL.Open("Postgres","postgres://postgres:postgres@host.com")```

here, the first argument to ```SQL.Open``` is the driver name (which will help the SQL package to identify which driver is to be used. We have already registered a driver with this name just above) and the second is the actual DSN.

Now to use DynamicSQl, check the following code

``` 
package main

import (
    "database/SQL"
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

Here, you must register your actual driver (i.e. Postgres/MySQL) with dynamicSQl before opening any connection.
Then while DB.Open(), the first argument must be "dynamicsql". The second argument must bbe a URI.
The scheme should be "dynamicsql". The host can be anything. Then the path should be always the actual path of file containing the latest DSN of the DB. 


