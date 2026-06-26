module github.com/tinta/identity

go 1.22

require (
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.2
	github.com/rs/zerolog v1.32.0
	github.com/tinta/shared v0.0.0
	golang.org/x/crypto v0.18.0
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/tinta/shared => ../../shared
