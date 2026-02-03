package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Quote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Text      string             `bson:"text"`
	Source    string             `bson:"source,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
}

func QuotesCollection() *mongo.Collection {
	return Database.Collection("quotes")
}

// GetAllQuotes returns all quotes from the database
func GetAllQuotes() ([]Quote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := QuotesCollection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotes []Quote
	if err := cursor.All(ctx, &quotes); err != nil {
		return nil, err
	}
	return quotes, nil
}

// GetRandomQuote returns a random quote using MongoDB aggregation
func GetRandomQuote() (*Quote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
	}

	cursor, err := QuotesCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotes []Quote
	if err := cursor.All(ctx, &quotes); err != nil {
		return nil, err
	}

	if len(quotes) == 0 {
		return nil, nil
	}
	return &quotes[0], nil
}

// AddQuote inserts a new quote
func AddQuote(text, source string) (*Quote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	quote := Quote{
		Text:      text,
		Source:    source,
		CreatedAt: time.Now(),
	}

	result, err := QuotesCollection().InsertOne(ctx, quote)
	if err != nil {
		return nil, err
	}

	quote.ID = result.InsertedID.(primitive.ObjectID)
	return &quote, nil
}

// DeleteQuote removes a quote by ID
func DeleteQuote(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := QuotesCollection().DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// CountQuotes returns the number of quotes
func CountQuotes() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return QuotesCollection().CountDocuments(ctx, bson.M{})
}
