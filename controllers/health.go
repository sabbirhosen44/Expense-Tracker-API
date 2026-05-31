package controllers

type HealthController struct {
	BaseController
}

// Get sends a 200 OK response to show the server is alive.
func (c *HealthController) GetHealthStatus() {
	c.respondOK("Server is running!", nil)
}
