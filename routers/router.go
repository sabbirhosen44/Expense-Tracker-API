package routers

import (
	"expense-tracker/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/api/v1/health", &controllers.HealthController{}, "get:Get")
}
