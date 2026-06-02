package routers

import (
	"expense-tracker/controllers"
	"expense-tracker/middlewares"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {

	ns := beego.NewNamespace("/api/v1",

		// Health Route
		beego.NSInclude(
			&controllers.HealthController{},
		),
		beego.NSRouter("/health", &controllers.HealthController{}, "get:GetHealthStatus"),

		// Auth Route Group
		beego.NSNamespace("/auth",
			beego.NSInclude(
				&controllers.AuthController{},
			),
			beego.NSRouter("/register", &controllers.AuthController{}, "post:Register"),
			beego.NSRouter("/login", &controllers.AuthController{}, "post:Login"),
		),

		// Expenses Route Group
		beego.NSNamespace("/expenses",
			beego.NSBefore(middlewares.AuthFilter),
			beego.NSInclude(
				&controllers.ExpenseController{},
			),

			beego.NSRouter("", &controllers.ExpenseController{}, "post:CreateExpense;get:ListExpenses"),
			beego.NSRouter("/:id", &controllers.ExpenseController{}, "get:GetExpense;put:UpdateExpense;delete:DeleteExpense"),
			beego.NSRouter("/summary", &controllers.ExpenseController{}, "get:GetSummary"),
		))

	beego.AddNamespace(ns)
}
