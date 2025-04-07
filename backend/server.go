package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors" // Importar el middleware de CORS
	analyzer "backend/analyzer" // Importar el paquete analyzer
)

type CommandRequest struct {
	Command string `json:"command"`
}

type CommandResponse struct {
	Output string `json:"output"`
}

func main() {
	// Crear una nueva instancia de Fiber
	app := fiber.New()

	// Configurar el middleware de CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000", // Permitir solicitudes desde el frontend
		AllowHeaders: "Origin, Content-Type, Accept", // Headers permitidos
		AllowMethods: "GET, POST, PUT, DELETE", // Métodos HTTP permitidos
	}))

	// Ruta de prueba
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("¡Hola, mundo!")
	})

	// Ruta para ejecutar comandos
	app.Post("/execute", func(c *fiber.Ctx) error {
		// Obtener el cuerpo de la solicitud
		var requestBody struct {
			Command string `json:"command"`
		}

		// Parsear el cuerpo de la solicitud
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(400).JSON(CommandResponse{
				Output: "Error: Petición inválida",
			})
		}

		commands := strings.Split(requestBody.Command, "\n")
		output := ""

		for _, cmd := range commands {
			if strings.TrimSpace(cmd) == "" || strings.HasPrefix(cmd, "#") {
				continue
			}

			result, err := analyzer.Analyzer(cmd)
			if err != nil {
				output += fmt.Sprintf("Error: %s\n", err.Error())
			} else {
				output += fmt.Sprintf("%s\n", result)
			}
		}

		if output == "" {
			output = "No se ejecutó ningún comando"
		}

		return c.JSON(CommandResponse{
			Output: output,
		})

	})

	// Iniciar el servidor en el puerto 8000
	log.Fatal(app.Listen(":8000"))
}