package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Subject struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Icon      string             `bson:"icon"`
	CreatedAt time.Time          `bson:"created_at"`
}

func SubjectsCollection() *mongo.Collection {
	return Database.Collection("subjects")
}

// GetAllSubjects returns all subjects
func GetAllSubjects() ([]Subject, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := SubjectsCollection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subjects []Subject
	if err := cursor.All(ctx, &subjects); err != nil {
		return nil, err
	}
	return subjects, nil
}

// GetSubjectByID returns a subject by its ID
func GetSubjectByID(id primitive.ObjectID) (*Subject, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var subject Subject
	err := SubjectsCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&subject)
	if err != nil {
		return nil, err
	}
	return &subject, nil
}

// AddSubject creates a new subject
func AddSubject(name, icon string) (*Subject, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	subject := Subject{
		Name:      name,
		Icon:      icon,
		CreatedAt: time.Now(),
	}

	result, err := SubjectsCollection().InsertOne(ctx, subject)
	if err != nil {
		return nil, err
	}

	subject.ID = result.InsertedID.(primitive.ObjectID)
	return &subject, nil
}

// AddSubjectIfNotExists creates a subject only if one with the same name doesn't exist
func AddSubjectIfNotExists(name, icon string) (*Subject, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if subject already exists
	var existing Subject
	err := SubjectsCollection().FindOne(ctx, bson.M{"name": name}).Decode(&existing)
	if err == nil {
		// Already exists
		return &existing, false, nil
	}
	if err != mongo.ErrNoDocuments {
		return nil, false, err
	}

	// Create new subject
	subject := Subject{
		Name:      name,
		Icon:      icon,
		CreatedAt: time.Now(),
	}

	result, err := SubjectsCollection().InsertOne(ctx, subject)
	if err != nil {
		return nil, false, err
	}

	subject.ID = result.InsertedID.(primitive.ObjectID)
	return &subject, true, nil
}

// DeleteSubject removes a subject by ID
func DeleteSubject(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := SubjectsCollection().DeleteOne(ctx, bson.M{"_id": id})
	return err
}
