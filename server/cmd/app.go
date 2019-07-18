package main

import (
	"fmt"
	"github.com/danibix95/FdP_tickets/server/internal/controller"
	"github.com/danibix95/FdP_tickets/server/internal/dbconn"
	"github.com/kataras/iris"
	"time"
)

func main() {
	controller.Greet()
	dbc := dbconn.New("logs")

	app := iris.Default()
	app.Get("/ping", func(ctx iris.Context) {

		dbc.PingDB() // test db connection

		_, _ = ctx.JSON(iris.Map{
			"message": fmt.Sprintf("Pong - %v", time.Now().Local()),
		})
	})

	// with _ = it is possible to ignore the value returned by the method
	_ = app.Run(iris.Addr(":8080"))
}
