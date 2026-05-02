package hat

// ComboConstraint represents one executable combo line from Freya analysis.
// Zones are plain strings matching the engine convention: "hand",
// "battlefield", "graveyard", "exile", "library", "command".
type ComboConstraint struct {
	Name            string              // human-readable combo name
	PiecesNeeded    []string            // card names required
	ZonesAccepted   map[string][]string // card name -> acceptable zones
	ManaRequired    int                 // total mana to execute the full sequence
	SequenceOrder   []string            // required cast/activate order
	NeedsProtection bool                // should we hold counterspell backup?
}

// ComboAssessment is the result of evaluating all combo lines for a seat.
type ComboAssessment struct {
	Executable   bool             // at least one combo can win this turn
	Assembling   bool             // within 1 piece + tutor available
	BestLine     *ComboConstraint // the most promising line
	NextAction   string           // card name to cast/activate next (if Executable)
	MissingPiece string           // what we need to tutor for (if Assembling)
	PiecesFound  int              // how many pieces of best line we have
	PiecesTotal  int              // how many the best line needs
}
