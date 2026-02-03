package main

import (
	"fmt"
	"log"

	"Beot/db"
)

var seedQuotes = []struct {
	Text     string
	Source   string
	Subjects []string // Empty = general quote for all subjects
}{
	// General quotes (shown for all subjects)
	{
		Text:   "Some of the greatest innovations have come from people who only succeeded because they were too dumb to know that what they were doing was impossible.",
		Source: "",
	},
	{
		Text:   "Game design is decision making, and decisions must be made with confidence.",
		Source: "",
	},
	{
		Text:   "A computer is a creative amplifier.",
		Source: "",
	},

	// Programming quotes (GoLang, React)
	{
		Text:     "First, solve the problem. Then, write the code.",
		Source:   "John Johnson",
		Subjects: []string{"GoLang", "React"},
	},
	{
		Text:     "Code is like humor. When you have to explain it, it's bad.",
		Source:   "Cory House",
		Subjects: []string{"GoLang", "React"},
	},
	{
		Text:     "Simplicity is the soul of efficiency.",
		Source:   "Austin Freeman",
		Subjects: []string{"GoLang", "React"},
	},
	{
		Text:     "Make it work, make it right, make it fast.",
		Source:   "Kent Beck",
		Subjects: []string{"GoLang", "React"},
	},
	{
		Text:     "The best error message is the one that never shows up.",
		Source:   "Thomas Fuchs",
		Subjects: []string{"GoLang", "React"},
	},

	// Music quotes
	{
		Text:     "Music is the shorthand of emotion.",
		Source:   "Leo Tolstoy",
		Subjects: []string{"Music"},
	},
	{
		Text:     "Without music, life would be a mistake.",
		Source:   "Friedrich Nietzsche",
		Subjects: []string{"Music"},
	},
	{
		Text:     "Music expresses that which cannot be put into words.",
		Source:   "Victor Hugo",
		Subjects: []string{"Music"},
	},
	{
		Text:     "If you aren't dropping, you aren't learning. And if you aren't learning, you aren't a juggler.",
		Source:   "Juggler's proverb",
		Subjects: []string{"Music"},
	},
	{
		Text:     "The only way to do great work is to love what you do.",
		Source:   "Steve Jobs",
		Subjects: []string{"Music"},
	},

	// Reading quotes
	{
		Text:     "A reader lives a thousand lives before he dies. The man who never reads lives only one.",
		Source:   "George R.R. Martin",
		Subjects: []string{"Reading"},
	},
	{
		Text:     "Reading is to the mind what exercise is to the body.",
		Source:   "Joseph Addison",
		Subjects: []string{"Reading"},
	},

	// Writing quotes
	{
		Text:     "There is nothing to writing. All you do is sit down at a typewriter and bleed.",
		Source:   "Ernest Hemingway",
		Subjects: []string{"Writing"},
	},
	{
		Text:     "Start writing, no matter what. The water does not flow until the faucet is turned on.",
		Source:   "Louis L'Amour",
		Subjects: []string{"Writing"},
	},
}

var seedSubjects = []struct {
	Name string
	Icon string
}{
	{Name: "GoLang", Icon: "🔷"},
	{Name: "React", Icon: "⚛"},
	{Name: "Music", Icon: "🎵"},
	{Name: "Reading", Icon: "📖"},
	{Name: "Writing", Icon: "✍"},
}

var seedPoems = []struct {
	OldEnglish    string
	ModernEnglish string
	Source        string
	LineRef       string
}{
	// The Wanderer passages
	{
		OldEnglish:    "Oft him ánhaga áre gebídeð,\nmetudes miltse, þéah þe hé módcearig",
		ModernEnglish: "Often the solitary one finds grace,\nthe Measurer's mercy, though he, anxious in heart",
		Source:        "The Wanderer",
		LineRef:       "lines 1-2",
	},
	{
		OldEnglish:    "Hwǽr cwóm mearg? Hwǽr cwóm mago?\nHwǽr cwóm máþþumgyfa?",
		ModernEnglish: "Where has the horse gone? Where has the man gone?\nWhere has the treasure-giver gone?",
		Source:        "The Wanderer",
		LineRef:       "lines 92-93",
	},
	{
		OldEnglish:    "Hwǽr cwóm symbla gesetu?\nHwǽr sindon seledréamas?",
		ModernEnglish: "Where have the seats of feasting gone?\nWhere are the joys of the hall?",
		Source:        "The Wanderer",
		LineRef:       "lines 93-94",
	},
	{
		OldEnglish:    "Éalá beorht bune! Éalá byrnwiga!\nÉalá þéodnes þrym!",
		ModernEnglish: "Alas, the bright cup! Alas, the mailed warrior!\nAlas, the glory of the prince!",
		Source:        "The Wanderer",
		LineRef:       "lines 94-95",
	},
	{
		OldEnglish:    "Til biþ se þe his tréowe gehealdeþ,\nne sceal nǽfre his torn tó rycene",
		ModernEnglish: "Good is he who keeps his faith,\nnor shall he ever too quickly show his grief",
		Source:        "The Wanderer",
		LineRef:       "lines 112-113",
	},
	{
		OldEnglish:    "Swá cwæð eardstapa, earfeþa gemyndig,\nwraþra wælsleahta, wine-mǽga hryre",
		ModernEnglish: "So spoke the earth-stepper, mindful of hardships,\nof cruel slaughters, the fall of kinsmen",
		Source:        "The Wanderer",
		LineRef:       "lines 6-7",
	},
	// Beowulf passages
	{
		OldEnglish:    "Hwæt! Wé Gár-Dena in géar-dagum,\nþéod-cyninga þrym gefrúnon",
		ModernEnglish: "Listen! We have heard of the glory\nof the Spear-Danes in days of old",
		Source:        "Beowulf",
		LineRef:       "lines 1-2",
	},
	{
		OldEnglish:    "Swá sceal geong guma góde gewyrcean,\nfromum feohgiftum on fæder bearme",
		ModernEnglish: "So should a young man do good deeds,\nwith rich gifts in his father's keeping",
		Source:        "Beowulf",
		LineRef:       "lines 20-21",
	},
	{
		OldEnglish:    "Wyrd oft nereð\nunfǽgne eorl, þonne his ellen déah",
		ModernEnglish: "Fate often saves\nan undoomed man, when his courage holds",
		Source:        "Beowulf",
		LineRef:       "lines 572-573",
	},
	{
		OldEnglish:    "Ure ǽghwylc sceal ende gebídan\nworolde lífes; wyrce sé þe móte\ndómes ǽr déaþe",
		ModernEnglish: "Each of us must await the end\nof worldly life; let him who may\nwin glory before death",
		Source:        "Beowulf",
		LineRef:       "lines 1386-1388",
	},
	{
		OldEnglish:    "Né bið swylc cwénlíc þéaw\nidese tó efnanne, þéah ðe híe ǽnlíc sý",
		ModernEnglish: "It is not queenly custom\nfor a woman to practice, though she be peerless",
		Source:        "Beowulf",
		LineRef:       "lines 1940-1941",
	},
	{
		OldEnglish:    "Sé þe his worde wéaldeð, wita manna gehwylc,\nwís on gewitte",
		ModernEnglish: "He who rules his words, every wise man,\nskilled in thought",
		Source:        "Beowulf",
		LineRef:       "lines 1705-1706",
	},
	{
		OldEnglish:    "Ic þæt þonne forhicge,\nswá mé Higelác síe, mín mondrihten,\nmódes blíðe",
		ModernEnglish: "I scorn therefore to carry sword or shield,\nif Hygelac, my liege lord,\nbe glad of heart",
		Source:        "Beowulf",
		LineRef:       "lines 435-437",
	},
	{
		OldEnglish:    "Nealles him on héape handgesteallan,\næðelinga bearn, ymbe gestódon\nhildecystum",
		ModernEnglish: "Not at all did the band of comrades,\nsons of nobles, stand about him\nwith battle valor",
		Source:        "Beowulf",
		LineRef:       "lines 2596-2598",
	},
}

func main() {
	fmt.Println("Connecting to MongoDB...")
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()

	fmt.Println("Seeding quotes...")

	for _, q := range seedQuotes {
		quote, err := db.AddQuoteWithSubjects(q.Text, q.Source, q.Subjects)
		if err != nil {
			log.Printf("Failed to add quote: %v", err)
			continue
		}
		subjectInfo := "general"
		if len(q.Subjects) > 0 {
			subjectInfo = fmt.Sprintf("%v", q.Subjects)
		}
		fmt.Printf("  Added [%s]: %s...\n", subjectInfo, truncate(quote.Text, 40))
	}

	count, _ := db.CountQuotes()
	fmt.Printf("Total quotes in database: %d\n", count)

	fmt.Println("\nSeeding subjects...")

	for _, s := range seedSubjects {
		subject, err := db.AddSubject(s.Name, s.Icon)
		if err != nil {
			log.Printf("Failed to add subject: %v", err)
			continue
		}
		fmt.Printf("  Added: %s %s\n", subject.Icon, subject.Name)
	}

	subjects, _ := db.GetAllSubjects()
	fmt.Printf("Total subjects in database: %d\n", len(subjects))

	fmt.Println("\nSeeding poems...")

	for _, p := range seedPoems {
		poem, err := db.AddPoem(p.OldEnglish, p.ModernEnglish, p.Source, p.LineRef)
		if err != nil {
			log.Printf("Failed to add poem: %v", err)
			continue
		}
		fmt.Printf("  Added: %s (%s)\n", poem.Source, poem.LineRef)
	}

	poemCount, _ := db.CountPoems()
	fmt.Printf("\nDone! Total poems in database: %d\n", poemCount)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
