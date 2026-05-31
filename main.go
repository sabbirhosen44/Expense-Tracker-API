package main

import (
	_ "expense-tracker/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	
	beego.Run()
}
