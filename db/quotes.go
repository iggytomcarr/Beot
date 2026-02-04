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
	Subjects  []string           `bson:"subjects,omitempty"` // Empty = general (shown for all)
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
	return GetRandomQuoteForSubject("")
}

// GetRandomQuoteForSubject returns a random quote for a specific subject
// It includes quotes tagged with that subject OR general quotes (no subjects)
func GetRandomQuoteForSubject(subjectName string) (*Quote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Build filter: quotes for this subject OR general quotes (empty/null subjects)
	var filter bson.M
	if subjectName != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"subjects": subjectName},
				{"subjects": bson.M{"$exists": false}},
				{"subjects": bson.M{"$size": 0}},
				{"subjects": nil},
			},
		}
	} else {
		filter = bson.M{}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
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

// AddQuote inserts a new quote (general, shown for all subjects)
func AddQuote(text, source string) (*Quote, error) {
	return AddQuoteWithSubjects(text, source, nil)
}

// AddQuoteWithSubjects inserts a new quote with subject tags
func AddQuoteWithSubjects(text, source string, subjects []string) (*Quote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	quote := Quote{
		Text:      text,
		Source:    source,
		Subjects:  subjects,
		CreatedAt: time.Now(),
	}

	result, err := QuotesCollection().InsertOne(ctx, quote)
	if err != nil {
		return nil, err
	}

	quote.ID = result.InsertedID.(primitive.ObjectID)
	return &quote, nil
}

// AddQuoteIfNotExists creates a quote only if one with the same text doesn't exist
func AddQuoteIfNotExists(text, source string, subjects []string) (*Quote, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if quote already exists
	var existing Quote
	err := QuotesCollection().FindOne(ctx, bson.M{"text": text}).Decode(&existing)
	if err == nil {
		return &existing, false, nil
	}
	if err != mongo.ErrNoDocuments {
		return nil, false, err
	}

	// Create new quote
	quote := Quote{
		Text:      text,
		Source:    source,
		Subjects:  subjects,
		CreatedAt: time.Now(),
	}

	result, err := QuotesCollection().InsertOne(ctx, quote)
	if err != nil {
		return nil, false, err
	}

	quote.ID = result.InsertedID.(primitive.ObjectID)
	return &quote, true, nil
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
