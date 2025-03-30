package common

import (
	"fmt"
	"math/rand"
	"time"
)

// List of words to use for box names
var words = []string{
	// Original 50 words
	"amazing", "brave", "clever", "daring", "eager", "fierce", "gentle", "happy",
	"intelligent", "joyful", "kind", "lively", "mighty", "noble", "peaceful", "quiet",
	"radiant", "swift", "tender", "unique", "valiant", "wise", "young", "zealous",
	"apple", "banana", "cherry", "dragon", "eagle", "falcon", "grape", "hawk",
	"iris", "jasmine", "kiwi", "lemon", "maple", "night", "ocean", "peach",
	"quail", "rose", "sun", "tiger", "unicorn", "violet", "wolf", "xylophone",
	"yacht", "zebra",
	// Additional 50 words
	"brilliant", "courageous", "determined", "energetic", "friendly", "graceful", "harmonious", "inspiring",
	"jovial", "knowledgeable", "luminous", "magnificent", "nurturing", "optimistic", "playful", "resilient",
	"serene", "thoughtful", "uplifting", "versatile", "wonderful", "xenial", "youthful", "zesty",
	"butterfly", "cloud", "dolphin", "elephant", "forest", "garden", "horizon", "island",
	"jungle", "koala", "lighthouse", "mountain", "nebula", "orchid", "penguin", "rainbow",
	"seagull", "thunder", "umbrella", "volcano", "waterfall", "xylophone", "yoga", "zinnia",
	// some cities, no spaces
	"tokyo", "paris", "london", "beijing", "moscow", "sydney", "mexico", "dubai", "sao", "mumbai", "seoul", "osaka", "cairo", "bangkok", "istanbul",
	"seattle", "santiago", "buenos", "warsaw", "vienna",
	"stockholm", "oslo", "copenhagen", "helsinki", "prague", "washington", "montreal", "toronto",
	"berlin", "madrid", "milan", "oslo", "paris", "rome", "stockholm", "vienna", "warsaw", "zagreb",
	// some famous IT companies
	"google", "apple", "facebook", "amazon", "microsoft", "ibm", "intel", "nvidia", "amd", "qualcomm",
	// some alcohol
	"absinthe", "brandy", "gin", "rum", "vodka", "whiskey", "wine", "beer", "sake", "tequila", "wine",
}

// GenerateBoxID creates a human-readable box ID in the format "word-word-XXX"
func GenerateBoxID() string {
	// Create a local random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Select two random words
	word1 := words[r.Intn(len(words))]
	word2 := words[r.Intn(len(words))]

	// Generate a random 4-digit number
	num := r.Intn(9000) + 1000

	// Format the ID
	return fmt.Sprintf("%s-%s-%04d", word1, word2, num)
}
