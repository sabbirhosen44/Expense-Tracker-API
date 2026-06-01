package routers

import (
	"expense-tracker/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {

	ns := beego.NewNamespace("/api/v1",
		beego.NSRouter("/health", &controllers.HealthController{}, "get:GetHealthStatus"),

		beego.NSNamespace("/auth",
			beego.NSRouter("/register", &controllers.AuthController{}, "post:Register"),
			beego.NSRouter("/login", &controllers.AuthController{}, "post:Login"),
		),

		beego.NSNamespace("/expenses",
			beego.NSRouter("", &controllers.ExpenseController{}, "post:CreateExpense;get:ListExpenses"),
			beego.NSRouter("/:id", &controllers.ExpenseController{}, "get:GetExpense;put:UpdateExpense;delete:DeleteExpense"),
		))

	beego.AddNamespace(ns)
}
