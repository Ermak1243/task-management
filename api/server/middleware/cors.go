package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Setup подключает CORS-мидлварь, разрешающий все origins,
// методы GET/POST/PUT/DELETE и стандартные заголовки.
func Setup(app *fiber.App) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Accept, Content-Type, Authorization",
	}))
}
