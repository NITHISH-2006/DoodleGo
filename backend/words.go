package main

import (
	"math/rand"
	"strings"
)

const (
	difficultyEasy   = "easy"
	difficultyMedium = "medium"
	difficultyHard   = "hard"
)

var easyWords = []string{
	"apple", "banana", "cat", "dog", "house", "tree", "car", "moon", "star", "book",
	"airplane", "anchor", "angel", "ant", "arrow", "baby", "bag", "ball", "balloon", "basket",
	"beach", "bear", "bed", "bee", "bell", "bicycle", "bird", "boat", "bone", "bottle",
	"bowl", "box", "bridge", "brush", "bucket", "bus", "butterfly", "button", "cake", "camera",
	"candle", "candy", "castle", "chair", "cheese", "cherry", "clock", "cloud", "clown", "coffee",
	"coin", "computer", "cookie", "crown", "cup", "diamond", "dinosaur", "door", "dragon", "drum",
	"duck", "egg", "elephant", "envelope", "eye", "feather", "fire", "fish", "flag", "flower",
	"football", "fork", "fox", "frog", "garden", "ghost", "gift", "glass", "glasses", "glove",
	"guitar", "hammer", "hat", "heart", "helicopter", "honey", "horse", "hospital", "ice", "island",
	"jacket", "juice", "kangaroo", "key", "kite", "ladder", "lamp", "leaf", "lemon", "lion",
	"lizard", "lock", "magnet", "map", "mask", "medal", "mirror", "monkey", "mountain", "mouse",
	"mushroom", "nail", "necklace", "needle", "nest", "nose", "ocean", "octopus", "onion", "orange",
	"owl", "paint", "panda", "paper", "parrot", "peach", "pear", "pencil", "penguin", "piano",
	"pie", "pig", "pillow", "pizza", "planet", "plant", "pocket", "popcorn", "potato", "present",
	"pumpkin", "rabbit", "rain", "rainbow", "robot", "rocket", "rose", "ruler", "sandwich", "scissors",
	"shark", "ship", "shoe", "skateboard", "skull", "snail", "snake", "snow", "snowman", "sock",
	"spider", "spoon", "squirrel", "strawberry", "sun", "sunglasses", "sword", "table", "teapot", "tent",
	"tiger", "toaster", "tomato", "tooth", "tower", "train", "truck", "turtle", "umbrella", "violin",
	"volcano", "watermelon", "whale", "wheel", "window", "wolf", "zebra", "zipper", "broom", "bat",
	"acorn", "barn", "bench", "blanket", "boot", "branch", "carrot", "chicken", "cow", "crab",
	"doll", "dolphin", "donkey", "donut", "dress",
}

var mediumWords = []string{
	"accordion", "alligator", "alpaca", "ambulance", "antelope", "aquarium", "armadillo", "armor", "atlas", "avocado",
	"axe", "badge", "bagel", "baklava", "balcony", "bandage", "banjo", "barometer", "battery", "beaver",
	"binoculars", "biscuit", "bison", "blender", "blimp", "blueprint", "bookshelf", "boomerang", "bracelet", "briefcase",
	"brownie", "buffalo", "bulldozer", "burrito", "cactus", "calculator", "calendar", "campfire", "cannon", "canoe",
	"cappuccino", "carousel", "cassette", "cauliflower", "cello", "chandelier", "cheeseburger", "chessboard", "chimpanzee", "chipmunk",
	"chocolate", "chopsticks", "clarinet", "cloak", "cobra", "compass", "coral", "coyote", "cradle", "crayon",
	"croissant", "cucumber", "cupcake", "cymbal", "dagger", "dartboard", "deer", "dentures", "dictionary", "diploma",
	"dishwasher", "domino", "doorknob", "dragonfly", "dumbbell", "dumpling", "easel", "eggplant", "escalator", "excavator",
	"falcon", "faucet", "ferret", "firework", "flamingo", "flute", "forklift", "fossil", "fountain", "galaxy",
	"gargoyle", "gazelle", "gecko", "geyser", "ginger", "glider", "globe", "goldfish", "goggles", "gondola",
	"gorilla", "gramophone", "grapefruit", "guacamole", "hammock", "hamster", "handcuffs", "harmonica", "harpoon", "hawk",
	"headphone", "hedgehog", "heron", "hexagon", "honeycomb", "hourglass", "hovercraft", "hummingbird", "hyena", "iguana",
	"jackhammer", "jaguar", "jellyfish", "jukebox", "kaleidoscope", "kayak", "kennel", "kettle", "keyboard", "kimono",
	"kingfisher", "kiwi", "koala", "ladle", "lagoon", "lasagna", "lawnmower", "lemur", "lemonade", "leopard",
	"lifeguard", "limousine", "llama", "locket", "loom", "lynx", "macaw", "magnifier", "mammoth", "manatee",
	"mannequin", "maracas", "mayonnaise", "meerkat", "megaphone", "melon", "metronome", "microscope", "microwave", "minecart",
	"minivan", "mole", "mongoose", "monocle", "moose", "motorcycle", "mousetrap", "muffin", "mural", "mustache",
	"narwhal", "newt", "newspaper", "noodle", "nozzle", "obelisk", "observatory", "omelet", "opossum", "orangutan",
	"orca", "origami", "ornament", "ostrich", "otter", "paddle", "pagoda", "paintbrush", "palette", "pancake",
	"papaya", "parachute", "parasol", "passport", "pasta", "pastry", "peacock", "pebble", "pelican", "pendulum",
	"penny", "periscope", "pheasant", "phone", "photo", "pickle", "picnic", "pier", "pineapple", "pinwheel",
}

var hardWords = []string{
	"abacus", "acceleration", "acoustics", "adaptation", "adventure", "advertisement", "aerodynamics", "agriculture", "algorithm", "altitude",
	"ambiguity", "anatomy", "anniversary", "anticipation", "aperture", "archaeology", "architecture", "archive", "arithmetic", "atmosphere",
	"astronomy", "athlete", "audience", "authority", "avalanche", "aviation", "balance", "banquet", "barrier", "behavior",
	"belief", "biodiversity", "biology", "blizzard", "boundary", "bravery", "breeze", "brightness", "broadcast", "bureaucracy",
	"calculation", "celebration", "ceremony", "challenge", "chaos", "chemistry", "civilization", "climate", "collaboration", "communication",
	"community", "comparison", "competition", "concentration", "conclusion", "conflict", "connection", "constellation", "contrast", "cooperation",
	"courage", "creativity", "curiosity", "database", "decision", "democracy", "destination", "determination", "discovery", "distance",
	"diversity", "dream", "drought", "ecology", "economy", "education", "electricity", "elevation", "emotion", "energy",
	"environment", "equality", "evolution", "exchange", "exercise", "experiment", "exploration", "expression", "failure", "fantasy",
	"fashion", "festival", "fiction", "focus", "force", "fortune", "freedom", "friendship", "friction", "future",
	"generation", "geography", "geometry", "gravity", "growth", "harmony", "heritage", "history", "hypothesis", "identity",
	"imagination", "impact", "infinity", "innovation", "inspiration", "instinct", "invention", "journey", "justice", "knowledge",
	"language", "liberty", "logic", "loneliness", "loyalty", "magic", "memory", "migration", "momentum", "motion",
	"mystery", "nature", "navigation", "network", "noise", "nostalgia", "observation", "opportunity", "orbit", "origin",
	"peace", "perspective", "philosophy", "physics", "pollution", "possibility", "power", "pressure", "progress", "promise",
	"protection", "quality", "quantity", "question", "radiation", "reaction", "reality", "reason", "reflection", "relationship",
	"religion", "rhythm", "risk", "rotation", "science", "season", "secret", "security", "sensation", "silence",
	"simplicity", "society", "solitude", "sound", "space", "speed", "stability", "strength", "structure", "survival",
	"symmetry", "technology", "temperature", "tension", "theory", "thought", "tradition", "transformation", "travel", "truth",
	"universe", "velocity", "victory", "vision", "volume", "voyage", "warmth", "weather", "weight", "wisdom",
	"wonder", "movement", "direction", "dimension", "frequency", "probability", "revolution", "sequence", "strategy", "vibration",
}

func wordsForDifficulty(difficulty string) []string {
	switch difficulty {
	case difficultyMedium:
		return mediumWords
	case difficultyHard:
		return hardWords
	default:
		return easyWords
	}
}

func newWordDeck(difficulty, lastWord string) []string {
	deck := append([]string(nil), wordsForDifficulty(difficulty)...)
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	if len(deck) > 1 && deck[0] == lastWord {
		deck[0], deck[1] = deck[1], deck[0]
	}
	return deck
}

func normalizeDifficulty(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case difficultyMedium:
		return difficultyMedium
	case difficultyHard:
		return difficultyHard
	default:
		return difficultyEasy
	}
}
