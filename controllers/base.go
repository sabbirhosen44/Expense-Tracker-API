package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
)

// APIResponse keeps the JSON keys in a clean, fixed order.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// BaseController lets other controllers reuse these response functions.
type BaseController struct {
	beego.Controller
}

// respondJSON sets the HTTP status and sends the final JSON data.
func (c *BaseController) respondJSON(status int, success bool, message string, data interface{}) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	c.ServeJSON()
}

// respondOK sends a 200 status when data is fetched or updated successfully.
func (c *BaseController) respondOK(message string, data interface{}) {
	c.respondJSON(200, true, message, data)
}

// respondCreated sends a 201 status when a new item is successfully made.
func (c *BaseController) respondCreated(message string, data interface{}) {
	c.respondJSON(201, true, message, data)
}

// respondBadRequest sends a 400 status when the user sends wrong or invalid data.
func (c *BaseController) respondBadRequest(message string) {
	c.respondJSON(400, false, message, nil)
}

// respondUnauthorized sends a 401 status when login tokens are missing or wrong.
func (c *BaseController) respondUnauthorized(message string) {
	c.respondJSON(401, false, message, nil)
}

// respondNotFound sends a 404 status when the requested item does not exist.
func (c *BaseController) respondNotFound(message string) {
	c.respondJSON(404, false, message, nil)
}

// respondConflict sends a 409 status when there is duplicate data, like an existing email.
func (c *BaseController) respondConflict(message string) {
	c.respondJSON(409, false, message, nil)
}

// respondInternalError sends a 500 status when the server or database crashes.
func (c *BaseController) respondInternalError(message string) {
	c.respondJSON(500, false, message, nil)
}
