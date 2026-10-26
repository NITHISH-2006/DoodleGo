package main

import "math/rand"

var missionDeck = []MissionInfo{
	{
		Id:    "picnic-day",
		Title: "Picnic Day",
		Brief: "Build the coziest picnic in the park.",
		Tasks: []string{"checkered blanket", "picnic basket", "sandwiches", "tall tree", "shining sun"},
	},
	{
		Id:    "space-picnic",
		Title: "Space Picnic",
		Brief: "Pack a lunch that is out of this world.",
		Tasks: []string{"rocket ship", "picnic moon", "astronaut", "floating sandwich", "tiny alien"},
	},
	{
		Id:    "underwater-party",
		Title: "Underwater Party",
		Brief: "Throw a bash at the bottom of the sea.",
		Tasks: []string{"coral castle", "party fish", "bubble stream", "treasure chest", "sea turtle"},
	},
	{
		Id:    "monster-mash",
		Title: "Monster Mash",
		Brief: "Invite the friendliest monsters to dance.",
		Tasks: []string{"friendly monster", "dance floor", "disco ball", "snack table", "party banner"},
	},
	{
		Id:    "jungle-gym",
		Title: "Jungle Gym",
		Brief: "Design a playground hidden in the jungle.",
		Tasks: []string{"hanging vine", "monkey bars", "parrot", "wooden swing", "leafy slide"},
	},
	{
		Id:    "beach-day",
		Title: "Beach Day",
		Brief: "Fill the shore with summer details.",
		Tasks: []string{"sandcastle", "beach umbrella", "seagull", "surfboard", "shell collection"},
	},
	{
		Id:    "winter-wonderland",
		Title: "Winter Wonderland",
		Brief: "Make a snowy scene with warm surprises.",
		Tasks: []string{"snowman", "pine trees", "ice skates", "hot cocoa", "snowy cabin"},
	},
	{
		Id:    "pizza-party",
		Title: "Pizza Party",
		Brief: "Build the ultimate pizza night.",
		Tasks: []string{"giant pizza", "pizza chef", "cheese pull", "soda bottles", "party hats"},
	},
	{
		Id:    "birthday-bash",
		Title: "Birthday Bash",
		Brief: "Decorate the best birthday table ever.",
		Tasks: []string{"layer cake", "wrapped presents", "balloons", "birthday banner", "party candles"},
	},
	{
		Id:    "robot-workshop",
		Title: "Robot Workshop",
		Brief: "Invent a helpful robot and its tools.",
		Tasks: []string{"robot body", "control panel", "spare gears", "tool rack", "charging station"},
	},
	{
		Id:    "haunted-house",
		Title: "Haunted House",
		Brief: "Draw a spooky home with playful ghosts.",
		Tasks: []string{"creaky door", "friendly ghost", "glowing window", "bats", "crooked fence"},
	},
	{
		Id:    "fairy-tale",
		Title: "Fairy Tale",
		Brief: "Illustrate a storybook adventure.",
		Tasks: []string{"storybook castle", "tiny fairy", "magic wand", "talking frog", "moonlit path"},
	},
	{
		Id:    "superhero-city",
		Title: "Superhero City",
		Brief: "Create a city that needs a hero.",
		Tasks: []string{"tall skyscraper", "flying hero", "comic burst", "hero cape", "rescued cat"},
	},
	{
		Id:    "farm-life",
		Title: "Farm Life",
		Brief: "Draw a busy morning on the farm.",
		Tasks: []string{"red barn", "tractor", "chickens", "hay bales", "farm dog"},
	},
	{
		Id:    "dinosaur-dig",
		Title: "Dinosaur Dig",
		Brief: "Set up a fossil hunting site.",
		Tasks: []string{"giant fossil", "dig tools", "paleontologist", "dinosaur footprint", "dusty jeep"},
	},
	{
		Id:    "music-festival",
		Title: "Music Festival",
		Brief: "Build a stage ready for a big show.",
		Tasks: []string{"concert stage", "drum kit", "spotlights", "crowd hands", "giant speakers"},
	},
	{
		Id:    "art-gallery",
		Title: "Art Gallery",
		Brief: "Hang a collection of surprising artwork.",
		Tasks: []string{"framed painting", "sculpture", "gallery bench", "tiny artist", "paint splatter"},
	},
	{
		Id:    "candy-shop",
		Title: "Candy Shop",
		Brief: "Stock the sweetest shop in town.",
		Tasks: []string{"candy jars", "lollipop stand", "shop counter", "striped awning", "chocolate bar"},
	},
	{
		Id:    "zoo-escape",
		Title: "Zoo Escape",
		Brief: "Draw a playful breakout at the zoo.",
		Tasks: []string{"open gate", "sneaky penguin", "zookeeper", "map sign", "getaway cart"},
	},
	{
		Id:    "desert-oasis",
		Title: "Desert Oasis",
		Brief: "Find a cool refuge in the hot sand.",
		Tasks: []string{"palm tree", "small pond", "camel", "sand dunes", "desert tent"},
	},
	{
		Id:    "mountain-climb",
		Title: "Mountain Climb",
		Brief: "Map out a daring climb to the peak.",
		Tasks: []string{"snowy peak", "climbing rope", "backpack", "mountain goat", "trail sign"},
	},
	{
		Id:    "campfire-night",
		Title: "Campfire Night",
		Brief: "Set up camp under a starry sky.",
		Tasks: []string{"crackling campfire", "tent", "sleeping bag", "marshmallow stick", "night sky"},
	},
	{
		Id:    "movie-night",
		Title: "Movie Night",
		Brief: "Create a cozy home cinema.",
		Tasks: []string{"big screen", "popcorn bowl", "comfy sofa", "movie projector", "remote control"},
	},
	{
		Id:    "breakfast-table",
		Title: "Breakfast Table",
		Brief: "Draw a breakfast worth waking up for.",
		Tasks: []string{"stack of pancakes", "coffee mug", "orange juice", "fried egg", "toast plate"},
	},
	{
		Id:    "tropical-island",
		Title: "Tropical Island",
		Brief: "Design a tiny island getaway.",
		Tasks: []string{"palm tree", "hammock", "volcano", "tropical bird", "message in a bottle"},
	},
	{
		Id:    "outer-space",
		Title: "Outer Space",
		Brief: "Explore a corner of the galaxy.",
		Tasks: []string{"ringed planet", "comet", "space station", "satellite", "shooting star"},
	},
	{
		Id:    "detective-office",
		Title: "Detective Office",
		Brief: "Set the scene for a mystery.",
		Tasks: []string{"magnifying glass", "case board", "trench coat", "notebook", "locked drawer"},
	},
	{
		Id:    "pirate-cove",
		Title: "Pirate Cove",
		Brief: "Hide a treasure along a secret coast.",
		Tasks: []string{"pirate ship", "treasure map", "parrot", "treasure chest", "skull flag"},
	},
	{
		Id:    "medieval-castle",
		Title: "Medieval Castle",
		Brief: "Draw a fortress fit for a royal court.",
		Tasks: []string{"castle tower", "drawbridge", "knight", "royal banner", "stone wall"},
	},
	{
		Id:    "future-city",
		Title: "Future City",
		Brief: "Imagine a city from a hundred years ahead.",
		Tasks: []string{"flying car", "glass tower", "robot helper", "sky bridge", "solar panel"},
	},
	{
		Id:    "garden-party",
		Title: "Garden Party",
		Brief: "Plan a cheerful party among the flowers.",
		Tasks: []string{"flower arch", "tea table", "butterflies", "garden lantern", "watering can"},
	},
	{
		Id:    "aquarium",
		Title: "Aquarium",
		Brief: "Fill a giant tank with ocean life.",
		Tasks: []string{"glass tank", "school of fish", "diver", "coral reef", "visitor window"},
	},
	{
		Id:    "train-station",
		Title: "Train Station",
		Brief: "Capture the rush of a busy platform.",
		Tasks: []string{"locomotive", "ticket booth", "platform clock", "suitcase", "station sign"},
	},
	{
		Id:    "airport",
		Title: "Airport",
		Brief: "Draw the scene before takeoff.",
		Tasks: []string{"airplane", "control tower", "runway", "passport", "luggage cart"},
	},
	{
		Id:    "soccer-match",
		Title: "Soccer Match",
		Brief: "Show the biggest moment of the game.",
		Tasks: []string{"goal net", "soccer ball", "cheering crowd", "referee whistle", "team flag"},
	},
	{
		Id:    "winter-sports",
		Title: "Winter Sports",
		Brief: "Create a snowy competition.",
		Tasks: []string{"skier", "pine slope", "medal podium", "snowboard", "finish line"},
	},
	{
		Id:    "circus",
		Title: "Circus",
		Brief: "Raise the tent for a spectacular show.",
		Tasks: []string{"circus tent", "juggler", "trapeze", "clown car", "popcorn stand"},
	},
	{
		Id:    "magic-school",
		Title: "Magic School",
		Brief: "Draw a classroom full of spells.",
		Tasks: []string{"floating book", "magic potion", "wizard hat", "spell circle", "owl perch"},
	},
	{
		Id:    "volcano-lab",
		Title: "Volcano Lab",
		Brief: "Build a safe place to study a volcano.",
		Tasks: []string{"erupting volcano", "lab table", "scientist", "lava sample", "warning sign"},
	},
	{
		Id:    "ocean-rescue",
		Title: "Ocean Rescue",
		Brief: "Draw a daring rescue at sea.",
		Tasks: []string{"rescue boat", "life ring", "lighthouse", "dolphin helper", "rough waves"},
	},
	{
		Id:    "dream-bedroom",
		Title: "Dream Bedroom",
		Brief: "Design a bedroom that feels like a dream.",
		Tasks: []string{"canopy bed", "glowing stars", "bookshelf", "fluffy rug", "window seat"},
	},
	{
		Id:    "street-market",
		Title: "Street Market",
		Brief: "Fill a lively street with colorful stalls.",
		Tasks: []string{"market stall", "fruit crates", "hanging lanterns", "shopping bag", "street musician"},
	},
}

func newMissionDeck() []MissionInfo {
	deck := append([]MissionInfo(nil), missionDeck...)
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck
}

func (r *Room) initMissionsLocked() {
	r.missionDeck = newMissionDeck()
	r.missionIndex = 0
	r.missionNumber = 1
	r.missionsCompleted = 0
	r.missionChecks = make([]bool, len(r.missionDeck[0].Tasks))
}

func (r *Room) missionStateLocked() *MissionStateData {
	if r.TogetherActivity != activityMissions || len(r.missionDeck) == 0 {
		return nil
	}
	if r.missionIndex < 0 || r.missionIndex >= len(r.missionDeck) {
		return nil
	}
	mission := r.missionDeck[r.missionIndex]
	return &MissionStateData{
		Mission: MissionInfo{
			Id:    mission.Id,
			Title: mission.Title,
			Brief: mission.Brief,
			Tasks: append([]string(nil), mission.Tasks...),
		},
		Checked:   append([]bool(nil), r.missionChecks...),
		Completed: r.missionsCompleted,
		Number:    r.missionNumber,
	}
}

func (r *Room) HandleMissionToggle(index int) {
	r.mu.Lock()
	if r.State != stateTogether || r.TogetherActivity != activityMissions || len(r.missionDeck) == 0 {
		r.mu.Unlock()
		return
	}
	if index < 0 || index >= len(r.missionChecks) {
		r.mu.Unlock()
		return
	}
	r.missionChecks[index] = !r.missionChecks[index]
	state := r.missionStateLocked()
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgMissionState, state))
}

func (r *Room) HandleNextMission() {
	r.mu.Lock()
	if r.State != stateTogether || r.TogetherActivity != activityMissions || len(r.missionDeck) == 0 {
		r.mu.Unlock()
		return
	}
	for _, checked := range r.missionChecks {
		if !checked {
			r.mu.Unlock()
			return
		}
	}

	r.missionsCompleted++
	r.missionIndex++
	if r.missionIndex >= len(r.missionDeck) {
		lastID := r.missionDeck[len(r.missionDeck)-1].Id
		r.missionDeck = newMissionDeck()
		if len(r.missionDeck) > 1 && r.missionDeck[0].Id == lastID {
			r.missionDeck[0], r.missionDeck[1] = r.missionDeck[1], r.missionDeck[0]
		}
		r.missionIndex = 0
	}
	r.missionNumber++
	r.missionChecks = make([]bool, len(r.missionDeck[r.missionIndex].Tasks))
	r.Strokes = nil
	state := r.missionStateLocked()
	r.mu.Unlock()

	r.Broadcast(newMessage(MsgMissionState, state))
	r.Broadcast(newMessage(MsgCanvasState, CanvasStateData{Strokes: []StrokeData{}}))
}
