package main

import (
	_ "github.com/231031/wellpaw-backend/doc"
	"github.com/231031/wellpaw-backend/internal/server"
)

// @title WellPaw API
// @version 1.0
// @description WellPaw API is a RESTful API for managing and tracking pet health
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:50001
// @BasePath /api
// @schemes http
func main() {
	server.Run()
}
