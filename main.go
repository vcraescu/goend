package main

import (
	"github.com/gocql/gocql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
	"log"
	"net/http"
	"os"
)

var carTable *table.Table
var session gocqlx.Session

type Car struct {
	ID         gocql.UUID `json:"id"`
	Identifier string     `json:"identifier"`
	Lat        string     `json:"lat"`
	Long       string     `json:"long"`
	Status     string     `json:"status"`
}

func basicCreateAndPopulateKeyspace() {
	err := session.ExecStmt(`CREATE KEYSPACE IF NOT EXISTS data WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`)
	if err != nil {
		log.Fatal("create keyspace:", err)
	}

	err = session.ExecStmt(`CREATE TABLE IF NOT EXISTS data.car (
		id uuid PRIMARY KEY,
		identifier text,
		lat text,
		long text,
		status text)`)
	if err != nil {
		log.Fatal("create table:", err)
	}
}

func mustParseUUID(s string) gocql.UUID {
	u, err := gocql.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

func initCassandraSession(hosts ...string) {
	var err error
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "data"
	session, err = gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		log.Fatal(err)
	}
}

func initTables() {
	if carTable == nil {
		carTable = table.New(table.Metadata{
			Name:    "car",
			Columns: []string{"id", "identifier", "lat", "long", "status"},
			PartKey: []string{"id"},
			SortKey: []string{"status"},
		})
	}
}

func getPostData(c echo.Context) error {
	car := new(Car)
	if err := c.Bind(car); err != nil {
		return err
	}

	car.ID = mustParseUUID(uuid.NewV4().String())
	q := session.Query(carTable.Insert()).BindStruct(car)
	if err := q.ExecRelease(); err != nil {
		c.Logger().Fatal(err)

		return err
	}

	return c.JSON(http.StatusOK, car)
}

func main() {
	initCassandraSession(os.Args[1:]...)
	basicCreateAndPopulateKeyspace()
	initTables()
	defer session.Close()

	e := echo.New()
	e.Use(middleware.Logger())
	e.POST("/", getPostData)

	e.Logger.Fatal(e.Start(":80"))
}
