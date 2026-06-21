module github.com/tinta/community

go 1.22

require (
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.2
	github.com/rs/zerolog v1.32.0
	github.com/tinta/shared v0.0.0
)

replace github.com/tinta/shared => ../../shared
