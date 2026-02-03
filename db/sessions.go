package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SessionStatus string

const (
	StatusCompleted SessionStatus = "completed"
	StatusAbandoned SessionStatus = "abandoned"
)

type Session struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	SubjectID   primitive.ObjectID `bson:"subject_id"`
	SubjectName string             `bson:"subject_name"` // Denormalized for easy display
	Duration    int                `bson:"duration"`     // In minutes
	Status      SessionStatus      `bson:"status"`
	StartedAt   time.Time          `bson:"started_at"`
	CompletedAt time.Time          `bson:"completed_at,omitempty"`
}

func SessionsCollection() *mongo.Collection {
	return Database.Collection("sessions")
}

// CreateSession saves a new session
func CreateSession(subjectID primitive.ObjectID, subjectName string, duration int, status SessionStatus, startedAt time.Time) (*Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session := Session{
		SubjectID:   subjectID,
		SubjectName: subjectName,
		Duration:    duration,
		Status:      status,
		StartedAt:   startedAt,
		CompletedAt: time.Now(),
	}

	result, err := SessionsCollection().InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}

	session.ID = result.InsertedID.(primitive.ObjectID)
	return &session, nil
}

// GetRecentSessions returns the most recent sessions
func GetRecentSessions(limit int) ([]Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "completed_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := SessionsCollection().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []Session
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetSessionStats returns statistics about sessions
type SessionStats struct {
	TotalSessions     int
	CompletedSessions int
	AbandonedSessions int
	TotalMinutes      int
	CurrentStreak     int
	LongestStreak     int
}

func GetSessionStats() (*SessionStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats := &SessionStats{}

	// Count total sessions
	total, err := SessionsCollection().CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	stats.TotalSessions = int(total)

	// Count completed sessions
	completed, err := SessionsCollection().CountDocuments(ctx, bson.M{"status": StatusCompleted})
	if err != nil {
		return nil, err
	}
	stats.CompletedSessions = int(completed)
	stats.AbandonedSessions = stats.TotalSessions - stats.CompletedSessions

	// Sum total minutes from completed sessions
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "status", Value: StatusCompleted}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$duration"}}},
		}}},
	}

	cursor, err := SessionsCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if len(results) > 0 {
		if total, ok := results[0]["total"].(int32); ok {
			stats.TotalMinutes = int(total)
		} else if total, ok := results[0]["total"].(int64); ok {
			stats.TotalMinutes = int(total)
		}
	}

	// Calculate streaks
	stats.CurrentStreak, stats.LongestStreak = calculateStreaks(ctx)

	return stats, nil
}

// calculateStreaks determines current and longest streaks
func calculateStreaks(ctx context.Context) (current, longest int) {
	// Get all completed sessions, sorted by date descending
	opts := options.Find().SetSort(bson.D{{Key: "completed_at", Value: -1}})
	cursor, err := SessionsCollection().Find(ctx, bson.M{"status": StatusCompleted}, opts)
	if err != nil {
		return 0, 0
	}
	defer cursor.Close(ctx)

	var sessions []Session
	if err := cursor.All(ctx, &sessions); err != nil {
		return 0, 0
	}

	if len(sessions) == 0 {
		return 0, 0
	}

	// Track unique days with completed sessions
	days := make(map[string]bool)
	for _, s := range sessions {
		dayKey := s.CompletedAt.Format("2006-01-02")
		days[dayKey] = true
	}

	// Convert to sorted slice of dates
	var sortedDays []time.Time
	for dayStr := range days {
		t, _ := time.Parse("2006-01-02", dayStr)
		sortedDays = append(sortedDays, t)
	}

	// Sort descending (most recent first)
	for i := 0; i < len(sortedDays)-1; i++ {
		for j := i + 1; j < len(sortedDays); j++ {
			if sortedDays[j].After(sortedDays[i]) {
				sortedDays[i], sortedDays[j] = sortedDays[j], sortedDays[i]
			}
		}
	}

	// Calculate current streak (from today or yesterday)
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	currentStreak := 0
	if len(sortedDays) > 0 {
		mostRecent := sortedDays[0].Truncate(24 * time.Hour)
		if mostRecent.Equal(today) || mostRecent.Equal(yesterday) {
			currentStreak = 1
			for i := 1; i < len(sortedDays); i++ {
				prev := sortedDays[i-1].Truncate(24 * time.Hour)
				curr := sortedDays[i].Truncate(24 * time.Hour)
				diff := prev.Sub(curr).Hours() / 24
				if diff == 1 {
					currentStreak++
				} else {
					break
				}
			}
		}
	}

	// Calculate longest streak
	longestStreak := 0
	if len(sortedDays) > 0 {
		streak := 1
		for i := 1; i < len(sortedDays); i++ {
			prev := sortedDays[i-1].Truncate(24 * time.Hour)
			curr := sortedDays[i].Truncate(24 * time.Hour)
			diff := prev.Sub(curr).Hours() / 24
			if diff == 1 {
				streak++
			} else {
				if streak > longestStreak {
					longestStreak = streak
				}
				streak = 1
			}
		}
		if streak > longestStreak {
			longestStreak = streak
		}
	}

	return currentStreak, longestStreak
}

// GetSessionsBySubject returns session counts per subject
func GetSessionsBySubject() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "status", Value: StatusCompleted}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$subject_name"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := SessionsCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	results := make(map[string]int)
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		if name, ok := result["_id"].(string); ok {
			if count, ok := result["count"].(int32); ok {
				results[name] = int(count)
			}
		}
	}

	return results, nil
}
