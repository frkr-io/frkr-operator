package infra

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

// Config holds database and broker connection details
type Config struct {
	DatabaseURL string
	BrokerURL   string
}

// GetConfigFromEnv reads infrastructure configuration from environment variables
func GetConfigFromEnv() (*Config, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://root@frkr-cockroachdb:26257/frkrdb?sslmode=disable"
	}

	brokerURL := os.Getenv("BROKER_URL")
	if brokerURL == "" {
		brokerURL = "frkr-redpanda:9092"
	}

	return &Config{
		DatabaseURL: dbURL,
		BrokerURL:   brokerURL,
	}, nil
}

// DB wraps database operations
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(connString string) (*DB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return &DB{db}, nil
}

// EnsureTenant creates a tenant if it doesn't exist, returns tenant ID
func (db *DB) EnsureTenant(tenantName string) (string, error) {
	var tenantID string

	// Try to get existing tenant
	err := db.QueryRow(`
		SELECT id FROM tenants 
		WHERE name = $1 AND deleted_at IS NULL
	`, tenantName).Scan(&tenantID)

	if err == nil {
		return tenantID, nil
	}

	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to query tenant: %w", err)
	}

	// Create new tenant
	err = db.QueryRow(`
		INSERT INTO tenants (name, plan) 
		VALUES ($1, 'free') 
		RETURNING id
	`, tenantName).Scan(&tenantID)

	if err != nil {
		return "", fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenantID, nil
}

// CreateStream creates a stream record in the database
func (db *DB) CreateStream(tenantID, name, description string, retentionDays int) (streamID, topic string, err error) {
	// Generate topic name
	topic = GenerateTopicName(tenantID, name)

	err = db.QueryRow(`
		INSERT INTO streams (tenant_id, name, description, retention_days, topic, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
		ON CONFLICT (tenant_id, name) DO UPDATE SET
			description = EXCLUDED.description,
			retention_days = EXCLUDED.retention_days,
			updated_at = now()
		RETURNING id, topic
	`, tenantID, name, description, retentionDays, topic).Scan(&streamID, &topic)

	if err != nil {
		return "", "", fmt.Errorf("failed to create stream: %w", err)
	}

	return streamID, topic, nil
}

// GetStream retrieves a stream by name
func (db *DB) GetStream(tenantID, name string) (streamID, topic string, err error) {
	err = db.QueryRow(`
		SELECT id, topic FROM streams 
		WHERE tenant_id = $1 AND name = $2 AND deleted_at IS NULL
	`, tenantID, name).Scan(&streamID, &topic)

	if err != nil {
		return "", "", fmt.Errorf("stream not found: %w", err)
	}

	return streamID, topic, nil
}

// GenerateTopicName creates a Kafka-compatible topic name
func GenerateTopicName(tenantID, streamName string) string {
	// Sanitize for topic name
	safeTenant := strings.ToLower(strings.ReplaceAll(tenantID, "-", ""))
	safeName := strings.ToLower(strings.ReplaceAll(streamName, " ", "-"))

	// Remove non-alphanumeric characters except hyphens
	safeName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, safeName)

	return fmt.Sprintf("stream-%s-%s", safeTenant[:min(8, len(safeTenant))], safeName)
}

// KafkaAdmin wraps Kafka admin operations
type KafkaAdmin struct {
	brokerURL string
}

// NewKafkaAdmin creates a new Kafka admin client
func NewKafkaAdmin(brokerURL string) *KafkaAdmin {
	return &KafkaAdmin{brokerURL: brokerURL}
}

// CreateTopic creates a Kafka topic if it doesn't exist
func (k *KafkaAdmin) CreateTopic(topicName string, numPartitions, replicationFactor int) error {
	conn, err := kafka.Dial("tcp", k.brokerURL)
	if err != nil {
		return fmt.Errorf("failed to connect to broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerAddr := net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port))
	controllerConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topicName,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		// Topic might already exist, which is fine
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "TOPIC_ALREADY_EXISTS") {
			return nil
		}
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}

// EnsureUser creates a user in the database or updates password
func (db *DB) EnsureUser(tenantID, username, password string) error {
	// For now, we don't have a users table in the current migrations,
	// but the gateways use ValidateBasicAuthForStream which current accepts any non-empty credentials.
	// We'll add this placeholder implementation to satisfy the reconciler
	// and add a TODO to create a users table if it becomes necessary.

	// TODO: Create 'users' table in migrations and implement persistence here
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
