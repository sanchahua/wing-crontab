package main

import (
	"app"
	"library/path"
	"controllers/consul"
)

func main() {
	app.Init(path.CurrentPath + "/config")
	defer app.Release()

	ctx := app.NewContext()
	control := consul.NewConsulController(ctx)
	defer control.Close()

	select {
		case <- ctx.Done():
	}
}
