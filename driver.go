package dynamicsql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type DynamicSQLDriver struct {
	OriginalSqlDriver driver.Driver
	DsnPool           map[string]*WrappedConn
	lock              *sync.RWMutex
}
type WrappedConn struct {
	newDsn    string
	erlierDsn string
	watcher   *fsnotify.Watcher
}

// Registers dynamicsql driver with sql package. Also defines underlying driver to be used i.e. passed as dri
func RegisterDriver(dri driver.Driver) {
	for _, val := range sql.Drivers() {
		if val == "dynamicsql" {
			panic("dynamicsql driver already registered. Please ensure that the driver is registered only once")
		}
	}
	sql.Register("dynamicsql", DynamicSQLDriver{
		OriginalSqlDriver: dri,
		DsnPool:           map[string]*WrappedConn{},
		lock:              &sync.RWMutex{},
	})
}

// Wrapper around conn.Open(dsn string)
func (dri DynamicSQLDriver) Open(identifier string) (driver.Conn, error) {
	log.Println("attempting via dynamicsql. identifier : ", identifier)

	path := ""
	if url, err := url.Parse(identifier); err != nil {
		return nil, err
	} else {
		if url.Scheme != "dynamicsql" {
			return nil, errors.New("Incorrect driver name in the identifier scheme : " + url.Scheme)
		}
		path = url.Path
	}

	dsn, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if val, ok := dri.DsnPool[identifier]; ok {

		conn, err := dri.OriginalSqlDriver.Open(val.newDsn)
		if err != nil && val.newDsn != val.erlierDsn {
			// this is optimistically done to try with the erlier dsn once.,
			// hoping the new dsn has been created but not applied yet to the DB
			return dri.OriginalSqlDriver.Open(val.erlierDsn)
		}
		return conn, err
	} else {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
		dri.lock.Lock()
		dri.DsnPool[identifier] = &WrappedConn{
			newDsn:    string(dsn),
			erlierDsn: string(dsn),
			watcher:   watcher,
		}

		err = dri.DsnPool[identifier].watcher.Add(path)
		if err != nil {
			dri.lock.Unlock()

			return nil, err
		}

		dri.lock.Unlock()

		dri.DsnPool[identifier].KeepWatching(path)

		return dri.OriginalSqlDriver.Open(string(dsn))
	}

}

// This is to keep a watch on file havign DSN. If the contents are changed, the new DSN is updated so that any new connection usses it
func (p *WrappedConn) KeepWatching(path string) {
	go func(path string) {
		for {
			select {
			case event, ok := <-p.watcher.Events:
				if !ok {
					log.Println("Error: events channel closed! This is abnormal")
					return
				}
				if event.Has(fsnotify.Write) {
					log.Println("state changed")
					p.erlierDsn = p.newDsn

					dsn, err := os.ReadFile(path)
					if err != nil {
						log.Println("Error: ", err)
						continue
					}

					p.newDsn = string(dsn)

				}
			case <-p.watcher.Errors:
				return
			}
		}
	}(path)

}
