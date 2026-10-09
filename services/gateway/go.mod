module github.com/aashishrajdev/halomail/services/gateway

go 1.25.1

require (
	github.com/aashishrajdev/halomail/services/contact v0.0.0
	github.com/aashishrajdev/halomail/services/identity v0.0.0
	github.com/aashishrajdev/halomail/services/notification v0.0.0
	github.com/aashishrajdev/halomail/services/scheduling v0.0.0
	github.com/aashishrajdev/halomail/services/shared v0.0.0
	github.com/aashishrajdev/halomail/services/template v0.0.0
)

require (
	cloud.google.com/go/compute/metadata v0.5.2 // indirect
	connectrpc.com/connect v1.18.1 // indirect
	connectrpc.com/otelconnect v0.7.2 // indirect
	github.com/aws/aws-sdk-go-v2 v1.47.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.21 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.20.8 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager v0.4.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.11.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.114.2 // indirect
	github.com/aws/smithy-go v1.28.4 // indirect
	github.com/caarlos0/env/v11 v11.3.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/redis/go-redis/v9 v9.7.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/grpc v1.69.4 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
)

replace github.com/aashishrajdev/halomail/services/shared => ../shared

replace github.com/aashishrajdev/halomail/services/identity => ../identity

replace github.com/aashishrajdev/halomail/services/scheduling => ../scheduling

replace github.com/aashishrajdev/halomail/services/contact => ../contact

replace github.com/aashishrajdev/halomail/services/template => ../template

replace github.com/aashishrajdev/halomail/services/notification => ../notification
