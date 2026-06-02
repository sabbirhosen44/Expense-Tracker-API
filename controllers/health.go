package controllers

type HealthController struct {
	BaseController
}

// @Summary Health Check
// @Description Check API status
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} controllers.APIResponse
// @router /health [get]
func (c *HealthController) GetHealthStatus() {
	c.respondOK("Server is running!", nil)
}
