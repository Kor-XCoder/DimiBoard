package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/storage/redis/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Use(helmet.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://dimiboard.coder.ac"},
		AllowCredentials: true,
	}))

	app.Use(logger.New())
	app.Use(compress.New())

	storage := redis.New(redis.Config{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "",
	})

	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		postgresDSN = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, postgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS notices (
                id UUID PRIMARY KEY,
                title TEXT NOT NULL,
                type TEXT NOT NULL,
                deadline TEXT
        )`)
	if err != nil {
		log.Fatalf("failed to ensure notices table: %v", err)
	}

	type Notice struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Type     string `json:"type"`
		Deadline string `json:"deadline"`
	}

	type Locations struct {
		Room     []int `json:"room"`
		After    []int `json:"after"`
		Club     []int `json:"club"`
		Restroom []int `json:"restroom"`
		Out      []int `json:"out"`
		Other    []int `json:"etc"`
	}

	app.Route("/notice").
		Get(func(c fiber.Ctx) error {
			var result []Notice
			rows, err := db.Query(ctx, "SELECT id, title, type, deadline FROM notices")
			if err != nil {
				return err
			}
			defer rows.Close()

			for rows.Next() {
				var notice Notice
				if err := rows.Scan(&notice.ID, &notice.Title, &notice.Type, &notice.Deadline); err != nil {
					return err
				}
				result = append(result, notice)
			}
			if err := rows.Err(); err != nil {
				return err
			}
			return c.JSON(result)
		}).
		Post(func(c fiber.Ctx) error {
			notice := new(Notice)
			if err := c.Bind().Body(notice); err != nil {
				return err
			}
			notice.ID = uuid.New().String()
			_, err := db.Exec(ctx, "INSERT INTO notices (id, title, type, deadline) VALUES ($1, $2, $3, $4)", notice.ID, notice.Title, notice.Type, notice.Deadline)
			if err != nil {
				return err
			}
			return c.SendStatus(fiber.StatusCreated)
		}).
		Delete(func(c fiber.Ctx) error {
			data := new(struct {
				ID string `json:"id"`
			})
			err := c.Bind().Body(data)
			if err != nil {
				return err
			}
			_, err = db.Exec(ctx, "DELETE FROM notices WHERE id = $1", data.ID)
			if err != nil {
				return err
			}
			return c.SendStatus(fiber.StatusNoContent)
		})
	app.Route("/location").
		Get(func(c fiber.Ctx) error {
			var result Locations
			locationMap := map[string]*[]int{
				"room":     &result.Room,
				"after":    &result.After,
				"club":     &result.Club,
				"restroom": &result.Restroom,
				"out":      &result.Out,
				"etc":      &result.Other,
			}
			for i := range 30 {
				location, err := storage.Get(fmt.Sprintf("location-%d", i+1))
				if err != nil {
					continue
				}

				locationStr := string(location)
				if slice, exists := locationMap[locationStr]; exists {
					*slice = append(*slice, i+1)
				}
			}
			return c.JSON(result)
		}).
		Post(func(c fiber.Ctx) error {
			for i := range 30 {
				storage.Set(fmt.Sprintf("location-%d", i+1), []byte("room"), 0)
			}
			return c.SendStatus(fiber.StatusCreated)
		}).
		Patch(func(c fiber.Ctx) error {
			data := new(struct {
				ID       string `json:"id"`
				Location string `json:"location"`
			})
			if err := c.Bind().Body(data); err != nil {
				return err
			}
			err := storage.Delete(fmt.Sprintf("location-%s", data.Location))
			if err != nil {
				return err
			}
			err = storage.Set(fmt.Sprintf("location-%s", data.ID), []byte(data.Location), 0)
			if err != nil {
				return err
			}
			return c.SendStatus(fiber.StatusNoContent)
		})
	log.Infof("Server starting on port %d", 8082)
	log.Fatal(app.Listen(":8082", fiber.ListenConfig{
		EnablePrefork: true,
	}))
}
