module github.com/philjestin/boatman-ecosystem/platform

go 1.24.1

require (
	github.com/nats-io/nats-server/v2 v2.12.4
	github.com/nats-io/nats.go v1.49.0
	github.com/philjestin/boatman-ecosystem/harness v0.0.0-00010101000000-000000000000
	github.com/philjestin/boatman-ecosystem/shared v0.0.0-00010101000000-000000000000
	modernc.org/sqlite v1.46.1
)

require (
	github.com/antithesishq/antithesis-sdk-go v0.5.0-default-no-op // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.4-0.20251030100505-070ab1a87a76 // indirect
	github.com/nats-io/jwt/v2 v2.8.0 // indirect
	github.com/nats-io/nkeys v0.4.12 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	modernc.org/libc v1.67.6 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

replace (
	github.com/philjestin/boatman-ecosystem/harness => ../harness
	github.com/philjestin/boatman-ecosystem/shared => ../shared
)
