package db

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

const (
	DefaultDatabase = "beot"
)

func init() {
	// Load .env file if it exists (silently ignore if not found)
	godotenv.Load()
}

// getMongoURI returns the MongoDB URI from environment variable
func getMongoURI() (string, error) {
	uri := os.Getenv("BEOT_MONGODB_URI")
	if uri == "" {
		return "", errors.New("BEOT_MONGODB_URI environment variable is required. Create a .env file or set it in your environment")
	}
	return uri, nil
}

// Connect establishes the MongoDB connection
func Connect() error {
	uri, err := getMongoURI()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	Client = client
	Database = client.Database(DefaultDatabase)
	return nil
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return Client.Disconnect(ctx)
	}
	return nil
}
