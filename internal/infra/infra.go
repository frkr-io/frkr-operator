package infra

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	commondb "github.com/frkr-io/frkr-common/db"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

// InfraConfig holds database and broker connection details
type InfraConfig struct {
	DatabaseURL string
	BrokerURL   string
}

// GetConfigFromEnv reads infrastructure configuration from environment variables
func GetConfigFromEnv() (*InfraConfig, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://root@frkr-cockroachdb:26257/frkrdb?sslmode=disable"
	}

	brokerURL := os.Getenv("BROKER_URL")
	if brokerURL == "" {
		brokerURL = "frkr-redpanda:9092"
	}

	return &InfraConfig{
		DatabaseURL: dbURL,
		BrokerURL:   brokerURL,
	}, nil
}

// DB wraps database operations
type DB struct {
	*sql.DB
}

// ConnectInfraDB creates a new database connection
func ConnectInfraDB(connString string) (*DB, error) {
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
	tenant, err := commondb.CreateOrGetTenant(db.DB, tenantName)
	if err != nil {
		return "", err
	}
	return tenant.ID, nil
}

// CreateStream creates a stream record in the database
func (db *DB) CreateStream(tenantID, name, description string, retentionDays int) (streamID, topic string, err error) {
	stream, err := commondb.CreateStream(db.DB, tenantID, name, description, retentionDays)
	if err != nil {
		return "", "", err
	}
	return stream.ID, stream.Topic, nil
}

// GetStream retrieves a stream by name
func (db *DB) GetStream(tenantID, name string) (streamID, topic string, err error) {
	stream, err := commondb.GetStream(db.DB, tenantID, name)
	if err != nil {
		return "", "", err
	}
	return stream.ID, stream.Topic, nil
}

// GenerateTopicName delegates to common implementation
func GenerateTopicName(tenantID, streamName string) string {
	return commondb.GenerateTopicName(tenantID, streamName)
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

// EnsureUser creates a user in the database
// Note: This logic was previously placeholder, now using shared implementation
func (db *DB) EnsureUser(tenantID, username, password string) error {
	_, err := commondb.CreateUser(db.DB, tenantID, username, password)
	return err
}
