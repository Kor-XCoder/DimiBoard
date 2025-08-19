package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/storage/redis/v3"
)

func main() {
	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Use(helmet.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:8081"},
		AllowCredentials: true,
	}))

	app.Use(logger.New())
	app.Use(compress.New())

	storage := redis.New(redis.Config{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "",
	})

	type Notice struct {
		Title    string `json:"title"`
		Type     string `json:"type"`
		Deadline string `json:"deadline"`
	}

	app.Route("/notice").
		Get(func(c fiber.Ctx) error {
			var result []Notice
			keys, err := storage.Keys()
			if err != nil {
				return err
			}

			for _, key := range keys {
				if strings.HasPrefix(string(key), "notice-") {
					value := new(Notice)
					raw, err := storage.Get(string(key))
					if err != nil {
						return err
					}
					err = sonic.Unmarshal(raw, value)
					if err != nil {
						return err
					}
					result = append(result, Notice{
						Title:    value.Title,
						Type:     value.Type,
						Deadline: value.Deadline,
					})
				}
			}
			return c.JSON(result)
		}).
		Post(func(c fiber.Ctx) error {
			notice := new(Notice)
			if err := c.Bind().Body(notice); err != nil {
				return err
			}
			data, err := sonic.Marshal(notice)
			if err != nil {
				return err
			}
			deadline, err := time.Parse("YYYY-MM-DD", notice.Deadline)
			if err != nil {
				return err
			}
			err = storage.Set(fmt.Sprintf("notice-%s", notice.Title), data, time.Until(deadline))
			if err != nil {
				return err
			}
			return c.JSON(notice)
		})
	log.Infof("Server starting on port %d", 8082)
	log.Fatal(app.Listen(":8082", fiber.ListenConfig{
		EnablePrefork: true,
	}))
}
