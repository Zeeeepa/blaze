// Package index implements an inverted index for full-text search
//
// ═══════════════════════════════════════════════════════════════════════════════
// WHAT IS AN INVERTED INDEX?
// ═══════════════════════════════════════════════════════════════════════════════
// An inverted index is like the index at the back of a book, but for search engines.
//
// Example: Given these documents:
//   Doc 1: "the quick brown fox"
//   Doc 2: "the lazy dog"
//   Doc 3: "quick brown dogs"
//
// The inverted index would look like:
//   "quick"  → [Doc1:Pos1, Doc3:Pos0]
//   "brown"  → [Doc1:Pos2, Doc3:Pos1]
//   "fox"    → [Doc1:Pos3]
//   "lazy"   → [Doc2:Pos1]
//   "dog"    → [Doc2:Pos2]
//   "dogs"   → [Doc3:Pos2]
//
// This allows us to:
// 1. Find documents containing a word instantly (without scanning all docs)
// 2. Find phrases by checking if word positions are consecutive
// 3. Rank results by how close words appear to each other (proximity)
//
// ═══════════════════════════════════════════════════════════════════════════════

package blaze

import (
	"errors"
	"log/slog"
	"sync"
)

// ═══════════════════════════════════════════════════════════════════════════════
// ERROR DEFINITIONS
// ═══════════════════════════════════════════════════════════════════════════════
// We define errors as package-level variables so they can be compared with ==
// This is a Go best practice for error handling.
var (
	ErrNoPostingList = errors.New("no posting list exists for token")
	ErrNoNextElement = errors.New("no next element found")
	ErrNoPrevElement = errors.New("no previous element found")
)

// ═══════════════════════════════════════════════════════════════════════════════
// CORE DATA STRUCTURE: InvertedIndex
// ═══════════════════════════════════════════════════════════════════════════════
// The InvertedIndex is the main structure that holds all searchable data.
//
// Architecture:
//
//	InvertedIndex
//	├── PostingsList: map[string]SkipList
//	│   ├── "quick" → SkipList of positions
//	│   ├── "brown" → SkipList of positions
//	│   └── "fox"   → SkipList of positions
//	└── mu: mutex for thread safety
//
// Why a map of SkipLists?
// - Map: Allows O(1) lookup of any word
// - SkipList: Allows O(log n) operations for finding positions (more on this in skip_list.go)
//
// Thread Safety:
// - Multiple goroutines might try to update the index simultaneously
// - The mutex (mu) ensures only one goroutine modifies the index at a time
// ═══════════════════════════════════════════════════════════════════════════════
type InvertedIndex struct {
	mu           sync.Mutex          // Protects against concurrent access
	PostingsList map[string]SkipList // Word → List of where it appears
}

// NewInvertedIndex creates a new empty inverted index
//
// Why use make() instead of {}?
// - make() pre-allocates memory for the map, which is slightly more efficient
// - For maps, both work the same, but make() is more explicit
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		PostingsList: make(map[string]SkipList),
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// INDEXING: Building the Search Index
// ═══════════════════════════════════════════════════════════════════════════════

// Index adds a document to the inverted index
//
// STEP-BY-STEP EXAMPLE:
// ----------------------
// Input: docID=1, document="The quick brown fox"
//
// Step 1: Tokenization
//
//	analyzer.Analyze() converts to: ["quick", "brown", "fox"]
//	(Note: "The" is removed as a stop word, and words are lowercased)
//
// Step 2: For each token, record its position
//
//	Token "quick" at position 0 in document 1
//	Token "brown" at position 1 in document 1
//	Token "fox"   at position 2 in document 1
//
// Step 3: Update the index
//
//	PostingsList["quick"] ← add Position{DocID:1, Offset:0}
//	PostingsList["brown"] ← add Position{DocID:1, Offset:1}
//	PostingsList["fox"]   ← add Position{DocID:1, Offset:2}
//
// Why record positions and not just document IDs?
// - Positions let us do phrase search ("brown fox" requires consecutive positions)
// - Positions let us rank by proximity (closer words = more relevant)
//
// Thread Safety Note:
// - We lock the entire indexing operation to prevent race conditions
// - If we didn't lock, two goroutines could corrupt the data structure
func (idx *InvertedIndex) Index(docID int, document string) {
	idx.mu.Lock()         // Acquire lock - only one goroutine can index at a time
	defer idx.mu.Unlock() // Release lock when function returns (even if it panics)

	slog.Info("indexing document", slog.Int("docID", docID))

	// STEP 1: Break document into searchable tokens
	// Example: "The Quick Brown Fox!" → ["quick", "brown", "fox"]
	tokens := Analyze(document)

	// STEP 2: Record each token's position in the document
	for position, token := range tokens {
		idx.indexToken(token, docID, position)
	}
}

// indexToken adds a single token occurrence to the index
//
// HOW IT WORKS:
// -------------
// 1. Check if we've seen this token before
//   - If yes: Get its existing SkipList
//   - If no:  Create a new SkipList for it
//
// 2. Insert the new position into the SkipList
//   - Position contains both document ID and word offset
//   - SkipList keeps positions sorted for fast searching
//
// 3. Store the updated SkipList back in the map
//
// DocumentID and Offset are stored as ints
// - The SkipList uses sentinel values (BOF=MinInt, EOF=MaxInt) to mark boundaries
// - All position values are integers (no casting needed)
func (idx *InvertedIndex) indexToken(token string, docID, position int) {
	// Check if this token already has a posting list
	skipList, exists := idx.getPostingList(token)
	if !exists {
		// First time seeing this token - create a new SkipList
		skipList = *NewSkipList()
	}

	// Add this occurrence to the token's posting list
	skipList.Insert(Position{
		DocumentID: docID,    // Which document?
		Offset:     position, // Where in the document?
	})

	// Save the updated SkipList back to the map
	// (In Go, maps don't update automatically when you modify a struct value)
	idx.PostingsList[token] = skipList
}

// getPostingList retrieves the posting list for a token
//
// This is a simple helper to avoid repeating map lookup code.
// Returns (skipList, true) if found, (empty, false) if not found.
func (idx *InvertedIndex) getPostingList(token string) (SkipList, bool) {
	skipList, exists := idx.PostingsList[token]
	return skipList, exists
}

// ═══════════════════════════════════════════════════════════════════════════════
// BASIC SEARCH OPERATIONS
// ═══════════════════════════════════════════════════════════════════════════════
// These four methods (First, Last, Next, Previous) form the foundation of
// all search operations. Everything else is built on top of these primitives.
//
// Think of them like iterator operations:
// - First: Go to the beginning
// - Last: Go to the end
// - Next: Move forward
// - Previous: Move backward
// ═══════════════════════════════════════════════════════════════════════════════

// First returns the first occurrence of a token in the index
//
// EXAMPLE:
// --------
// Given: "quick" appears at [Doc1:Pos1, Doc3:Pos0, Doc5:Pos2]
// First("quick") returns Doc3:Pos0 (the earliest occurrence)
//
// Use case: Start searching for a token from the beginning
func (idx *InvertedIndex) First(token string) (Position, error) {
	skipList, exists := idx.getPostingList(token)
	if !exists {
		return EOFDocument, ErrNoPostingList
	}

	// The first position is at the bottom level (level 0) of the SkipList
	// The Head node points to the first real node via Tower[0]
	return skipList.Head.Tower[0].Key, nil
}

// Last returns the last occurrence of a token in the index
//
// EXAMPLE:
// --------
// Given: "quick" appears at [Doc1:Pos1, Doc3:Pos0, Doc5:Pos2]
// Last("quick") returns Doc5:Pos2 (the latest occurrence)
//
// Use case: Search backwards from the end
func (idx *InvertedIndex) Last(token string) (Position, error) {
	skipList, exists := idx.getPostingList(token)
	if !exists {
		return EOFDocument, ErrNoPostingList
	}

	// Traverse to the end of the SkipList
	return skipList.Last(), nil
}

// Next finds the next occurrence of a token after the given position
//
// EXAMPLE:
// --------
// Given: "brown" appears at [Doc1:Pos2, Doc3:Pos1, Doc3:Pos5, Doc5:Pos0]
// Next("brown", Doc3:Pos1) returns Doc3:Pos5
// Next("brown", Doc3:Pos5) returns Doc5:Pos0
// Next("brown", Doc5:Pos0) returns EOF (no more occurrences)
//
// Special cases:
// - If currentPos is BOF (beginning of file), return First
// - If currentPos is already EOF (end of file), stay at EOF
//
// Use case: Iterate through all occurrences of a word
func (idx *InvertedIndex) Next(token string, currentPos Position) (Position, error) {
	// Special case: Starting from the beginning
	if currentPos.IsBeginning() {
		return idx.First(token)
	}

	// Special case: Already at the end
	if currentPos.IsEnd() {
		return EOFDocument, nil
	}

	// Get the posting list for this token
	skipList, exists := idx.getPostingList(token)
	if !exists {
		return EOFDocument, ErrNoPostingList
	}

	// Find the next position after currentPos in the SkipList
	// FindGreaterThan returns the smallest position > currentPos
	nextPos, _ := skipList.FindGreaterThan(currentPos)
	return nextPos, nil
}

// Previous finds the previous occurrence of a token before the given position
//
// EXAMPLE:
// --------
// Given: "brown" appears at [Doc1:Pos2, Doc3:Pos1, Doc3:Pos5, Doc5:Pos0]
// Previous("brown", Doc5:Pos0) returns Doc3:Pos5
// Previous("brown", Doc3:Pos5) returns Doc3:Pos1
// Previous("brown", Doc1:Pos2) returns BOF (no earlier occurrences)
//
// Use case: Search backwards through occurrences
func (idx *InvertedIndex) Previous(token string, currentPos Position) (Position, error) {
	// Special case: Starting from the end
	if currentPos.IsEnd() {
		return idx.Last(token)
	}

	// Special case: Already at the beginning
	if currentPos.IsBeginning() {
		return BOFDocument, nil
	}

	// Get the posting list for this token
	skipList, exists := idx.getPostingList(token)
	if !exists {
		return BOFDocument, ErrNoPostingList
	}

	// Find the previous position before currentPos in the SkipList
	// FindLessThan returns the largest position < currentPos
	prevPos, _ := skipList.FindLessThan(currentPos)
	return prevPos, nil
}
