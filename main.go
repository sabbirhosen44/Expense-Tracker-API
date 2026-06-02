package main

import (
	_ "expense-tracker/routers"
	"net/http"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	beego.SetStaticPath("/swagger", "swagger")
	beego.Handler("/swagger", http.RedirectHandler("/swagger/index.html", http.StatusFound))

	beego.Run()
}
