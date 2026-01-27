# Lesson 4: MongoDB Integration

In this lesson you'll connect to MongoDB and create the data layer for persisting sessions, quotes, and subjects.

---

## What You'll Learn

- Connecting to MongoDB from Go
- The repository pattern for data access
- Creating models with BSON tags
- CRUD operations (Create, Read, Update, Delete)
- Seeding default data

---

## Prerequisites

Make sure MongoDB is running:

**Option A: Docker (recommended)**
```bash
docker run -d -p 27017:27017 --name beot-mongo mongo:latest
```

**Option B: Local installation**
```bash
# macOS
brew services start mongodb-community

# Windows - run as service or:
mongod
```

**Verify it's running:**
```bash
# If you have mongosh installed:
mongosh --eval "db.version()"
```

---

## Project Structure

Expand your structure:

```
beot/
├── main.go
├── ui/
│   ├── app.go
│   ├── menu.go
│   ├── timer.go
│   └── styles.go
├── db/
│   ├── mongo.go      # Connection
│   ├── quotes.go     # Quote repository
│   ├── sessions.go   # Session repository
│   └── subjects.go   # Subject repository
└── models/
    ├── quote.go
    ├── session.go
    └── subject.go
```

```bash
mkdir db models
```

---

## Exercise 4.1: Data Models

Create `models/quote.go`:

```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Quote represents a motivational quote
type Quote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Text      string             `bson:"text"`
	Source    string             `bson:"source,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
}
```

Create `models/subject.go`:

```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Subject represents a focus area
type Subject struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Icon      string             `bson:"icon"`
	Color     string             `bson:"color"`
	CreatedAt time.Time          `bson:"created_at"`
}

// DefaultSubjects are seeded on first run
var DefaultSubjects = []Subject{
	{Name: "GoLang", Icon: "🐹", Color: "#00ADD8"},
	{Name: "Godot", Icon: "🎮", Color: "#478CBF"},
	{Name: "React", Icon: "⚛️", Color: "#61DAFB"},
	{Name: "Music", Icon: "🎹", Color: "#9B59B6"},
	{Name: "Reading", Icon: "📚", Color: "#F39C12"},
	{Name: "General", Icon: "⭐", Color: "#ECF0F1"},
}
```

Create `models/session.go`:

```go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SessionStatus indicates how a session ended
type SessionStatus string

const (
	StatusCompleted SessionStatus = "completed"
	StatusAbandoned SessionStatus = "abandoned"
)

// Session represents a pomodoro session
type Session struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	SubjectID   primitive.ObjectID `bson:"subject_id"`
	SubjectName string             `bson:"subject_name"` // Denormalized for easy display
	SubjectIcon string             `bson:"subject_icon"` // Denormalized for easy display
	Duration    int                `bson:"duration"`     // In minutes
	Status      SessionStatus      `bson:"status"`
	StartedAt   time.Time          `bson:"started_at"`
	CompletedAt time.Time          `bson:"completed_at"`
}
```

**Understanding BSON tags:**

```go
`bson:"field_name"`         // Maps Go field to MongoDB field
`bson:"_id,omitempty"`      // MongoDB's ID field, omit if empty
`bson:"source,omitempty"`   // Omit from document if empty string
```

---

## Exercise 4.2: Database Connection

Create `db/mongo.go`:

```go
package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection names
const (
	QuotesCollection   = "quotes"
	SessionsCollection = "sessions"
	SubjectsCollection = "subjects"
)

// DB wraps the MongoDB client and database
type DB struct {
	client   *mongo.Client
	database *mongo.Database
}

// Connect establishes a connection to MongoDB
func Connect(uri, dbName string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return &DB{
		client:   client,
		database: client.Database(dbName),
	}, nil
}

// Disconnect closes the MongoDB connection
func (db *DB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.client.Disconnect(ctx)
}

// Collection returns a collection by name
func (db *DB) Collection(name string) *mongo.Collection {
	return db.database.Collection(name)
}
```

**Key concepts:**

1. **Context with timeout:** Prevents hanging forever on connection issues
2. **Ping:** Verifies the connection actually works
3. **Error wrapping:** `fmt.Errorf("message: %w", err)` preserves the original error

---

## Exercise 4.3: Quote Repository

Create `db/quotes.go`:

```go
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"Beot/models"
)

// QuoteRepo handles quote database operations
type QuoteRepo struct {
	db *DB
}

// NewQuoteRepo creates a new quote repository
func NewQuoteRepo(db *DB) *QuoteRepo {
	return &QuoteRepo{db: db}
}

// Create adds a new quote
func (r *QuoteRepo) Create(ctx context.Context, text, source string) (*models.Quote, error) {
	quote := models.Quote{
		Text:      text,
		Source:    source,
		CreatedAt: time.Now(),
	}

	result, err := r.db.Collection(QuotesCollection).InsertOne(ctx, quote)
	if err != nil {
		return nil, err
	}

	quote.ID = result.InsertedID.(primitive.ObjectID)
	return &quote, nil
}

// GetAll returns all quotes
func (r *QuoteRepo) GetAll(ctx context.Context) ([]models.Quote, error) {
	cursor, err := r.db.Collection(QuotesCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotes []models.Quote
	if err := cursor.All(ctx, &quotes); err != nil {
		return nil, err
	}

	return quotes, nil
}

// GetRandom returns a random quote using MongoDB's $sample
func (r *QuoteRepo) GetRandom(ctx context.Context) (*models.Quote, error) {
	pipeline := bson.A{
		bson.M{"$sample": bson.M{"size": 1}},
	}

	cursor, err := r.db.Collection(QuotesCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotes []models.Quote
	if err := cursor.All(ctx, &quotes); err != nil {
		return nil, err
	}

	if len(quotes) == 0 {
		return nil, nil // No quotes in database
	}

	return &quotes[0], nil
}

// Delete removes a quote by ID
func (r *QuoteRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(QuotesCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Count returns the number of quotes
func (r *QuoteRepo) Count(ctx context.Context) (int64, error) {
	return r.db.Collection(QuotesCollection).CountDocuments(ctx, bson.M{})
}

// SeedDefaults adds starter quotes if the collection is empty
func (r *QuoteRepo) SeedDefaults(ctx context.Context) error {
	count, err := r.Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Already has data
	}

	quotes := []interface{}{
		models.Quote{
			Text:      "Some of the greatest innovations have come from people who only succeeded because they were too dumb to know that what they were doing was impossible.",
			CreatedAt: time.Now(),
		},
		models.Quote{
			Text:      "Game design is decision making, and decisions must be made with confidence.",
			CreatedAt: time.Now(),
		},
		models.Quote{
			Text:      "If you aren't dropping, you aren't learning. And if you aren't learning, you aren't a juggler.",
			Source:    "Juggler's saying",
			CreatedAt: time.Now(),
		},
		models.Quote{
			Text:      "A computer is a creative amplifier.",
			CreatedAt: time.Now(),
		},
		models.Quote{
			Text:      "The secret to getting ahead is getting started.",
			Source:    "Mark Twain",
			CreatedAt: time.Now(),
		},
	}

	_, err = r.db.Collection(QuotesCollection).InsertMany(ctx, quotes)
	return err
}
```

---

## Exercise 4.4: Subject Repository

Create `db/subjects.go`:

```go
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"Beot/models"
)

// SubjectRepo handles subject database operations
type SubjectRepo struct {
	db *DB
}

// NewSubjectRepo creates a new subject repository
func NewSubjectRepo(db *DB) *SubjectRepo {
	return &SubjectRepo{db: db}
}

// Create adds a new subject
func (r *SubjectRepo) Create(ctx context.Context, name, icon, color string) (*models.Subject, error) {
	subject := models.Subject{
		Name:      name,
		Icon:      icon,
		Color:     color,
		CreatedAt: time.Now(),
	}

	result, err := r.db.Collection(SubjectsCollection).InsertOne(ctx, subject)
	if err != nil {
		return nil, err
	}

	subject.ID = result.InsertedID.(primitive.ObjectID)
	return &subject, nil
}

// GetAll returns all subjects
func (r *SubjectRepo) GetAll(ctx context.Context) ([]models.Subject, error) {
	cursor, err := r.db.Collection(SubjectsCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subjects []models.Subject
	if err := cursor.All(ctx, &subjects); err != nil {
		return nil, err
	}

	return subjects, nil
}

// GetByID returns a subject by its ID
func (r *SubjectRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Subject, error) {
	var subject models.Subject
	err := r.db.Collection(SubjectsCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&subject)
	if err != nil {
		return nil, err
	}
	return &subject, nil
}

// Delete removes a subject by ID
func (r *SubjectRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(SubjectsCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Count returns the number of subjects
func (r *SubjectRepo) Count(ctx context.Context) (int64, error) {
	return r.db.Collection(SubjectsCollection).CountDocuments(ctx, bson.M{})
}

// SeedDefaults adds starter subjects if the collection is empty
func (r *SubjectRepo) SeedDefaults(ctx context.Context) error {
	count, err := r.Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	subjects := make([]interface{}, len(models.DefaultSubjects))
	for i, s := range models.DefaultSubjects {
		s.CreatedAt = time.Now()
		subjects[i] = s
	}

	_, err = r.db.Collection(SubjectsCollection).InsertMany(ctx, subjects)
	return err
}
```

---

## Exercise 4.5: Session Repository

Create `db/sessions.go`:

```go
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"Beot/models"
)

// SessionRepo handles session database operations
type SessionRepo struct {
	db *DB
}

// NewSessionRepo creates a new session repository
func NewSessionRepo(db *DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// Create adds a new session
func (r *SessionRepo) Create(ctx context.Context, session *models.Session) error {
	result, err := r.db.Collection(SessionsCollection).InsertOne(ctx, session)
	if err != nil {
		return err
	}
	session.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetToday returns all sessions from today
func (r *SessionRepo) GetToday(ctx context.Context) ([]models.Session, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	cursor, err := r.db.Collection(SessionsCollection).Find(ctx,
		bson.M{"started_at": bson.M{"$gte": startOfDay}},
		options.Find().SetSort(bson.M{"started_at": 1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []models.Session
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetRecent returns the most recent sessions
func (r *SessionRepo) GetRecent(ctx context.Context, limit int) ([]models.Session, error) {
	cursor, err := r.db.Collection(SessionsCollection).Find(ctx,
		bson.M{},
		options.Find().SetSort(bson.M{"started_at": -1}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []models.Session
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetDaysWithCompletedSessions returns dates that have completed sessions
// Used for streak calculation
func (r *SessionRepo) GetDaysWithCompletedSessions(ctx context.Context, limit int) ([]time.Time, error) {
	pipeline := bson.A{
		// Only completed sessions count for streaks
		bson.M{"$match": bson.M{"status": models.StatusCompleted}},
		// Sort newest first
		bson.M{"$sort": bson.M{"started_at": -1}},
		// Group by date (year, month, day)
		bson.M{"$group": bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$started_at"},
				"month": bson.M{"$month": "$started_at"},
				"day":   bson.M{"$dayOfMonth": "$started_at"},
			},
		}},
		// Sort by date descending
		bson.M{"$sort": bson.M{"_id.year": -1, "_id.month": -1, "_id.day": -1}},
		// Limit results
		bson.M{"$limit": limit},
	}

	cursor, err := r.db.Collection(SessionsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID struct {
			Year  int `bson:"year"`
			Month int `bson:"month"`
			Day   int `bson:"day"`
		} `bson:"_id"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	days := make([]time.Time, len(results))
	for i, r := range results {
		days[i] = time.Date(r.ID.Year, time.Month(r.ID.Month), r.ID.Day, 0, 0, 0, 0, time.UTC)
	}

	return days, nil
}

// GetStats returns session statistics
func (r *SessionRepo) GetStats(ctx context.Context) (completed, abandoned int64, err error) {
	completed, err = r.db.Collection(SessionsCollection).CountDocuments(ctx,
		bson.M{"status": models.StatusCompleted})
	if err != nil {
		return 0, 0, err
	}

	abandoned, err = r.db.Collection(SessionsCollection).CountDocuments(ctx,
		bson.M{"status": models.StatusAbandoned})
	if err != nil {
		return 0, 0, err
	}

	return completed, abandoned, nil
}
```

---

## Exercise 4.6: Connecting in main.go

Update `main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"Beot/db"
	"Beot/ui"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.Connect("mongodb://localhost:27017", "beot")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		fmt.Println("Make sure MongoDB is running:")
		fmt.Println("  docker run -d -p 27017:27017 --name beot-mongo mongo:latest")
		os.Exit(1)
	}
	defer database.Disconnect()

	// Seed default data
	quoteRepo := db.NewQuoteRepo(database)
	subjectRepo := db.NewSubjectRepo(database)

	if err := quoteRepo.SeedDefaults(ctx); err != nil {
		fmt.Printf("Failed to seed quotes: %v\n", err)
	}

	if err := subjectRepo.SeedDefaults(ctx); err != nil {
		fmt.Printf("Failed to seed subjects: %v\n", err)
	}

	// Run the TUI
	p := tea.NewProgram(ui.NewAppModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
```

**Run it:**

```bash
go run main.go
```

---

## Verifying the Data

Open a MongoDB shell to check your data:

```bash
# Using mongosh
mongosh

# Or with Docker
docker exec -it beot-mongo mongosh
```

```javascript
use beot

// Check quotes
db.quotes.find().pretty()

// Check subjects
db.subjects.find().pretty()

// Count documents
db.quotes.countDocuments()
db.subjects.countDocuments()
```

---

## Understanding MongoDB Operations

### BSON Filters

```go
// Empty filter - matches all documents
bson.M{}

// Simple equality
bson.M{"status": "completed"}

// Comparison operators
bson.M{"started_at": bson.M{"$gte": startTime}}
bson.M{"duration": bson.M{"$lt": 30}}

// Multiple conditions (AND)
bson.M{
    "status": "completed",
    "subject_id": someID,
}

// OR conditions
bson.M{"$or": bson.A{
    bson.M{"status": "completed"},
    bson.M{"status": "abandoned"},
}}
```

### Find Options

```go
options.Find().
    SetSort(bson.M{"started_at": -1}).  // Sort descending
    SetLimit(10).                        // Max 10 results
    SetSkip(5)                           // Skip first 5
```

### Aggregation Pipelines

For complex queries, use aggregation:

```go
pipeline := bson.A{
    bson.M{"$match": bson.M{"status": "completed"}},
    bson.M{"$group": bson.M{
        "_id": "$subject_name",
        "count": bson.M{"$sum": 1},
    }},
    bson.M{"$sort": bson.M{"count": -1}},
}

cursor, err := collection.Aggregate(ctx, pipeline)
```

---

## Checkpoint Tasks

Before moving to Lesson 5, make sure you can:

- [ ] App connects to MongoDB on startup
- [ ] Default quotes are seeded
- [ ] Default subjects are seeded
- [ ] Check data in mongosh: `db.quotes.find()`
- [ ] Running again doesn't duplicate seed data
- [ ] **Challenge:** Add a `GetBySource` method to QuoteRepo
- [ ] **Challenge:** Add an `Update` method to SubjectRepo
- [ ] **Challenge:** Create an index on `sessions.started_at` for faster queries

---

## Common Gotchas

### "Connection refused"

MongoDB isn't running. Start it:

```bash
docker start beot-mongo
# or
docker run -d -p 27017:27017 --name beot-mongo mongo:latest
```

### "Context deadline exceeded"

The connection is timing out. Check:
1. MongoDB is running
2. Port 27017 is correct
3. No firewall blocking

### "Cannot find module"

Make sure your imports match your module name:

```go
import "Beot/db"      // Matches: module Beot
import "beot/db"      // Matches: module beot
```

### "Type assertion failed"

When casting InsertedID:

```go
// Safe way
if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
    quote.ID = oid
}
```

### "Cursor already closed"

Always defer cursor.Close() and don't use the cursor after:

```go
cursor, err := collection.Find(ctx, filter)
if err != nil {
    return nil, err
}
defer cursor.Close(ctx)

// Use cursor here, before the function returns
```

---

## What's Next

In **Lesson 5**, we'll:
- Add subject selection before starting a timer
- Save sessions to the database
- Display subjects with their icons and colours

The data layer is ready. Time to use it!

---

## Quick Reference

```go
// Connect
db, err := db.Connect("mongodb://localhost:27017", "dbname")
defer db.Disconnect()

// Insert
result, err := collection.InsertOne(ctx, document)
id := result.InsertedID.(primitive.ObjectID)

// Find one
var doc Model
err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)

// Find many
cursor, err := collection.Find(ctx, bson.M{})
defer cursor.Close(ctx)
var docs []Model
err = cursor.All(ctx, &docs)

// Update
_, err := collection.UpdateOne(ctx,
    bson.M{"_id": id},
    bson.M{"$set": bson.M{"field": "value"}},
)

// Delete
_, err := collection.DeleteOne(ctx, bson.M{"_id": id})

// Count
count, err := collection.CountDocuments(ctx, bson.M{})

// Aggregate
cursor, err := collection.Aggregate(ctx, pipeline)
```

---

*"A computer is a creative amplifier."*

Data persists. The foundation holds. Build upon it. 🗄️
