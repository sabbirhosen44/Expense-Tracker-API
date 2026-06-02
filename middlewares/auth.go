package middlewares

import (
	"expense-tracker/models"
	"net/http"
	"strconv"

	"github.com/beego/beego/v2/server/web/context"
)

func AuthFilter(ctx *context.Context) {
	header := ctx.Input.Header("X-User-ID")
	userID, err := strconv.Atoi(header)

	// Validate ID and check if user exists
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil || userID <= 0 {
		ctx.Output.SetStatus(http.StatusUnauthorized)
		ctx.Output.JSON(map[string]interface{}{
			"success": false,
			"message": "Unauthorized",
		}, true, false)
		ctx.Abort(http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Optionally store the userID in the context for the controller to use
	ctx.Input.SetData("userID", userID)
}
