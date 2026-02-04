package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Poem struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	OldEnglish    string             `bson:"old_english"`
	ModernEnglish string             `bson:"modern_english"`
	Source        string             `bson:"source"`
	LineRef       string             `bson:"line_ref,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"`
}

func PoemsCollection() *mongo.Collection {
	return Database.Collection("poems")
}

// GetAllPoems returns all poems from the database
func GetAllPoems() ([]Poem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := PoemsCollection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var poems []Poem
	if err := cursor.All(ctx, &poems); err != nil {
		return nil, err
	}
	return poems, nil
}

// GetRandomPoem returns a random poem using MongoDB aggregation
func GetRandomPoem() (*Poem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
	}

	cursor, err := PoemsCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var poems []Poem
	if err := cursor.All(ctx, &poems); err != nil {
		return nil, err
	}

	if len(poems) == 0 {
		return nil, nil
	}
	return &poems[0], nil
}

// AddPoem inserts a new poem passage
func AddPoem(oldEnglish, modernEnglish, source, lineRef string) (*Poem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	poem := Poem{
		OldEnglish:    oldEnglish,
		ModernEnglish: modernEnglish,
		Source:        source,
		LineRef:       lineRef,
		CreatedAt:     time.Now(),
	}

	result, err := PoemsCollection().InsertOne(ctx, poem)
	if err != nil {
		return nil, err
	}

	poem.ID = result.InsertedID.(primitive.ObjectID)
	return &poem, nil
}

// AddPoemIfNotExists creates a poem only if one with the same source and lineRef doesn't exist
func AddPoemIfNotExists(oldEnglish, modernEnglish, source, lineRef string) (*Poem, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if poem already exists
	var existing Poem
	err := PoemsCollection().FindOne(ctx, bson.M{"source": source, "line_ref": lineRef}).Decode(&existing)
	if err == nil {
		return &existing, false, nil
	}
	if err != mongo.ErrNoDocuments {
		return nil, false, err
	}

	// Create new poem
	poem := Poem{
		OldEnglish:    oldEnglish,
		ModernEnglish: modernEnglish,
		Source:        source,
		LineRef:       lineRef,
		CreatedAt:     time.Now(),
	}

	result, err := PoemsCollection().InsertOne(ctx, poem)
	if err != nil {
		return nil, false, err
	}

	poem.ID = result.InsertedID.(primitive.ObjectID)
	return &poem, true, nil
}

// DeletePoem removes a poem by ID
func DeletePoem(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := PoemsCollection().DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// CountPoems returns the number of poems
func CountPoems() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return PoemsCollection().CountDocuments(ctx, bson.M{})
}
