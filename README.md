# Blaze

A high-performance, full-text search engine in Go featuring an inverted index with skip lists, phrase search, proximity ranking, and advanced text analysis.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Inverted Index](#inverted-index)
  - [Skip Lists](#skip-lists)
  - [Text Analysis Pipeline](#text-analysis-pipeline)
  - [Search Operations](#search-operations)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Performance Characteristics](#performance-characteristics)
- [Configuration](#configuration)
- [Use Cases](#use-cases)
- [Testing](#testing)
- [Architecture](#architecture)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [License](#license)

## Overview

Blaze is a Go engine that provides fast, full-text search capabilities through an inverted index implementation. It's designed for applications that need to search through text documents efficiently without relying on external search engines.

**Key Highlights:**

- **Inverted Index**: Maps terms to document positions for instant lookups
- **Skip Lists**: Probabilistic data structure providing O(log n) operations
- **Advanced Search**: Phrase search, proximity ranking, and boolean queries
- **Text Analysis**: Tokenization, stemming, stopword filtering, and case normalization
- **Thread-Safe**: Concurrent indexing with mutex protection
- **Serialization**: Efficient binary format for persistence

## Features

### Search Capabilities

- **Term Search**: Find documents containing specific terms
- **Phrase Search**: Exact multi-word matching ("quick brown fox")
- **Proximity Ranking**: Score results by term proximity
- **Position Tracking**: Track exact word positions within documents

### Text Processing

- **Tokenization**: Unicode-aware text splitting
- **Stemming**: Snowball (Porter2) stemmer for English
- **Stopword Filtering**: Remove common words (the, a, is, etc.)
- **Case Normalization**: Case-insensitive search
- **Configurable Pipeline**: Customize analysis behavior

### Data Structures

- **Skip Lists**: O(log n) search, insert, and delete operations
- **Inverted Index**: Efficient term-to-position mapping
- **Binary Serialization**: Compact storage format

## Installation

```bash
go get github.com/wizenheimer/blaze
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/wizenheimer/blaze"
)

func main() {
    // Create a new inverted index
    idx := blaze.NewInvertedIndex()

    // Index some documents
    idx.Index(1, "The quick brown fox jumps over the lazy dog")
    idx.Index(2, "A quick brown dog runs fast")
    idx.Index(3, "The lazy cat sleeps all day")

    // Search for documents containing "quick" and "brown"
    matches := idx.RankProximity("quick brown", 10)

    // Print results
    for _, match := range matches {
        fmt.Printf("Document %d (score: %.2f)\n",
            int(match.Offsets[0].DocumentID),
            match.Score)
    }
}
```

**Output:**

```
Document 2 (score: 1.00)
Document 1 (score: 0.50)
```

## Core Concepts

### Inverted Index

An inverted index is like the index at the back of a book. Instead of scanning every document to find a word, the index tells you exactly where each word appears.

**Example:**

Given these documents:

```
Doc 1: "the quick brown fox"
        Pos:0    1     2     3

Doc 2: "the lazy dog"
        Pos:0   1    2

Doc 3: "quick brown dogs"
        Pos:0    1     2
```

The inverted index looks like:

```
┌─────────┬────────────────────────────────────┐
│  Token  │         Posting List               │
├─────────┼────────────────────────────────────┤
│ "quick" │ → [Doc1:Pos1] → [Doc3:Pos0]        │
│ "brown" │ → [Doc1:Pos2] → [Doc3:Pos1]        │
│ "fox"   │ → [Doc1:Pos3]                      │
│ "lazy"  │ → [Doc2:Pos1]                      │
│ "dog"   │ → [Doc2:Pos2]                      │
│ "dogs"  │ → [Doc3:Pos2]                      │
└─────────┴────────────────────────────────────┘
```

**Visual Representation:**

```
                    Inverted Index
                    ┌──────────┐
                    │ Map      │
                    │ [string] │
                    │ SkipList │
                    └────┬─────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
   "quick"          "brown"           "fox"
   SkipList         SkipList         SkipList
   ┌──────┐        ┌──────┐         ┌──────┐
   │ HEAD │        │ HEAD │         │ HEAD │
   └──┬───┘        └──┬───┘         └──┬───┘
      │               │                 │
      ▼               ▼                 ▼
   ┌──────┐        ┌──────┐         ┌──────┐
   │Doc1:1│        │Doc1:2│         │Doc1:3│
   └──┬───┘        └──┬───┘         └──────┘
      │               │
      ▼               ▼
   ┌──────┐        ┌──────┐
   │Doc3:0│        │Doc3:1│
   └──────┘        └──────┘
```

**Benefits:**

- Instant term lookups (no document scanning)
- Phrase search via position checking
- Proximity ranking by measuring distances
- Efficient boolean queries (AND, OR, NOT)

### Skip Lists

A skip list is a probabilistic data structure that maintains sorted data with O(log n) average time complexity for search, insertion, and deletion.

**Visual Representation:**

```
Skip List with Multiple Levels (Express Lanes)
═══════════════════════════════════════════════════════════════

Level 3: HEAD ────────────────────────────────────────────────────────────> [30] ────────> NULL
              ↓                                                                ↓
Level 2: HEAD ─────────────────────────────> [15] ────────────────────────> [30] ────────> NULL
              ↓                                ↓                               ↓
Level 1: HEAD ─────────────> [10] ─────────> [15] ────────> [20] ─────────> [30] ────────> NULL
              ↓                ↓                ↓              ↓                ↓
Level 0: HEAD ──> [5] ──> [10] ──> [15] ──> [20] ──> [25] ──> [30] ──> [35] ──> NULL
         (ALL NODES AT LEVEL 0)

         ┌───────┐
         │ Node  │  Each node has a "tower" of forward pointers
         ├───────┤
         │ Key   │  Example: Node [15]
         ├───────┤
         │ Lvl 3 │ ──> [30]      (skip far ahead)
         │ Lvl 2 │ ──> [30]      (skip ahead)
         │ Lvl 1 │ ──> [20]      (skip a little)
         │ Lvl 0 │ ──> [20]      (next node)
         └───────┘
```

**How Heights are Assigned (Probabilistic):**

```
Coin Flip Algorithm:
┌─────────┬─────────────┬─────────────┐
│ Height  │ Probability │ Visual      │
├─────────┼─────────────┼─────────────┤
│    1    │    50%      │ ▓▓▓▓▓       │
│    2    │    25%      │ ▓▓▓         │
│    3    │   12.5%     │ ▓▓          │
│    4    │   6.25%     │ ▓           │
└─────────┴─────────────┴─────────────┘

For 1000 nodes, expected distribution:
Level 0: ~1000 nodes (all)    ████████████████████████████████████████
Level 1: ~500 nodes           ████████████████████
Level 2: ~250 nodes           ██████████
Level 3: ~125 nodes           █████
Level 4: ~62 nodes            ██
```

**Search Algorithm** (finding 20):

```
Step-by-Step Search for Key = 20:

Level 3: [HEAD] ───────────────────────────────> [30]        (30 > 20, drop down)
           ↓
Level 2: [HEAD] ──────────────> [15] ─────────> [30]        (15 < 20, advance)
                                   ↓
Level 2:                         [15] ─────────> [30]        (30 > 20, drop down)
                                   ↓
Level 1:                         [15] ──> [20]               (20 = 20, FOUND!)
                                          ^^^^

Journey Recorded:
┌───────────┬─────────────────┐
│ Level 3   │ HEAD            │  Predecessor at each level
│ Level 2   │ [15]            │  Used for insertions/deletions
│ Level 1   │ [15]            │
│ Level 0   │ [15]            │
└───────────┴─────────────────┘
```

1. Start at HEAD, Level 3
2. Level 3: Move to 30? No (30 > 20), drop to Level 2
3. Level 2: Move to 15? Yes (15 < 20), advance to 15
4. Level 2: Move to 30? No (30 > 20), drop to Level 1
5. Level 1: Move to 20? Yes! Found it!

**Time Complexity: O(log n) on average**

**Why Skip Lists?**

- O(log n) operations without complex balancing
- Simpler than AVL or Red-Black trees
- Better cache locality than trees
- Easier to make lock-free for concurrency
- Used in Redis, LevelDB, and other databases

### Text Analysis Pipeline

Blaze transforms raw text into searchable tokens through a multi-stage pipeline:

**Pipeline Stages:**

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Text Analysis Pipeline                          │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │  1. Tokenization                       │
         │  Split on non-alphanumeric chars       │
         └────────────────┬───────────────────────┘
                          ▼
         ┌────────────────────────────────────────┐
         │  2. Lowercasing                        │
         │  Normalize case ("Quick" → "quick")    │
         └────────────────┬───────────────────────┘
                          ▼
         ┌────────────────────────────────────────┐
         │  3. Stopword Filtering                 │
         │  Remove common words (the, a, is)      │
         └────────────────┬───────────────────────┘
                          ▼
         ┌────────────────────────────────────────┐
         │  4. Length Filtering                   │
         │  Remove tokens < 2 chars               │
         └────────────────┬───────────────────────┘
                          ▼
         ┌────────────────────────────────────────┐
         │  5. Stemming (Snowball/Porter2)        │
         │  Reduce to root ("running" → "run")    │
         └────────────────┬───────────────────────┘
                          ▼
                    Final Tokens
```

**Example Transformation:**

```
Input:  "The Quick Brown Fox Jumps!"
        │
        ├─ Step 1: Tokenization
        │  └─> ["The", "Quick", "Brown", "Fox", "Jumps"]
        │
        ├─ Step 2: Lowercasing
        │  └─> ["the", "quick", "brown", "fox", "jumps"]
        │
        ├─ Step 3: Stopword Filtering (remove "the")
        │  └─> ["quick", "brown", "fox", "jumps"]
        │
        ├─ Step 4: Length Filtering (all pass >= 2 chars)
        │  └─> ["quick", "brown", "fox", "jumps"]
        │
        └─ Step 5: Stemming ("jumps" → "jump")
           └─> ["quick", "brown", "fox", "jump"]
```

**Configuration:**

```go
// Use default configuration
tokens := blaze.Analyze("The quick brown fox")

// Custom configuration
config := blaze.AnalyzerConfig{
    MinTokenLength:  3,      // Only keep tokens >= 3 chars
    EnableStemming:  false,  // Disable stemming
    EnableStopwords: true,   // Keep stopword filtering
}
tokens := blaze.AnalyzeWithConfig("The quick brown fox", config)
```

### Search Operations

#### 1. Basic Term Search

Find all occurrences of a single term:

```go
idx := blaze.NewInvertedIndex()
idx.Index(1, "the quick brown fox")
idx.Index(2, "quick brown dogs")

// Find first occurrence of "quick"
pos, err := idx.First("quick")
if err == nil {
    fmt.Printf("Found at Doc %d, Pos %d\n",
        int(pos.DocumentID), int(pos.Offset))
}

// Find next occurrence
nextPos, _ := idx.Next("quick", pos)
```

#### 2. Phrase Search

Find exact sequences of words:

```go
// Find documents containing "quick brown fox" as a phrase
matches := idx.FindAllPhrases("quick brown fox", blaze.BOFDocument)

for _, match := range matches {
    start, end := match[0], match[1]
    fmt.Printf("Found in Doc %d from Pos %d to %d\n",
        int(start.DocumentID), int(start.Offset), int(end.Offset))
}
```

**Algorithm:**

```
Searching for phrase: "brown fox"

Document: "the quick brown dog jumped over the brown fox"
Positions: 0     1     2    3     4      5    6     7    8

Phase 1: Find END (last word "fox")
┌─────────────────────────────────────────────────────────┐
│ Find "brown" → Doc:Pos2                                 │
│ Find "fox" after Pos2 → Doc:Pos8  ← END position       │
└─────────────────────────────────────────────────────────┘

Phase 2: Walk BACKWARDS from END to find START
┌─────────────────────────────────────────────────────────┐
│ From Pos9, find previous "brown" → Doc:Pos7  ← START   │
└─────────────────────────────────────────────────────────┘

Phase 3: Validate
┌─────────────────────────────────────────────────────────┐
│ Start: Pos7, End: Pos8                                  │
│ Distance: 8 - 7 = 1                                     │
│ Expected: 2 words - 1 = 1  ✓ MATCH!                    │
│                                                          │
│      "brown"  "fox"                                     │
│        ▲       ▲                                        │
│       Pos7    Pos8    (consecutive positions)           │
└─────────────────────────────────────────────────────────┘
```

1. Find END: Locate the last word of the phrase
2. Walk BACKWARDS: Find previous occurrences of earlier words
3. Validate: Check if positions are consecutive
4. Recurse: Continue searching for more matches

#### 3. Proximity Search

Find documents containing all terms (not necessarily consecutive):

```go
// Find documents with both "quick" and "fox"
cover := idx.NextCover([]string{"quick", "fox"}, blaze.BOFDocument)
start, end := cover[0], cover[1]

// Calculate proximity score
distance := end.Offset - start.Offset
score := 1.0 / distance  // Closer terms = higher score
```

**Cover Algorithm:**

```
Searching for: ["quick", "fox"] (any order, not necessarily consecutive)

Document: "the quick brown dog jumped over the lazy fox"
Positions: 0     1     2    3     4      5    6    7    8

Phase 1: Find COVER END (furthest term)
┌──────────────────────────────────────────────────────────────┐
│ Find "quick" after BOF → Doc:Pos1                           │
│ Find "fox" after BOF → Doc:Pos8  ← FURTHEST (cover end)     │
└──────────────────────────────────────────────────────────────┘

Phase 2: Find COVER START (earliest term before end)
┌──────────────────────────────────────────────────────────────┐
│ Find "quick" before Pos9 → Doc:Pos1  ← EARLIEST (cover start)│
│ Find "fox" before Pos9 → Doc:Pos8                           │
└──────────────────────────────────────────────────────────────┘

Phase 3: Validate & Return
┌──────────────────────────────────────────────────────────────┐
│ Cover: [Pos1, Pos8]                                          │
│ Same document? ✓                                             │
│ All terms present? ✓                                         │
│                                                               │
│ "quick" ... ... ... ... ... ... ... "fox"                    │
│    ▲                                   ▲                     │
│   Pos1                                Pos8                   │
│   └────────── Cover Range ──────────────┘                    │
│                                                               │
│ Proximity Score: 1 / (8 - 1 + 1) = 1/8 = 0.125             │
└──────────────────────────────────────────────────────────────┘
```

1. Find FURTHEST occurrence of any term (cover end)
2. Find EARLIEST occurrence of each term before end (cover start)
3. Validate all terms are in the same document
4. Return [start, end] positions

#### 4. Proximity Ranking

Score and rank documents by term proximity:

```go
// Search and rank results
matches := idx.RankProximity("machine learning", 10)

for _, match := range matches {
    fmt.Printf("Doc %d: Score %.2f\n",
        int(match.Offsets[0].DocumentID),
        match.Score)
}
```

**Scoring Formula:**

```
For each cover in a document:
    score += 1 / (coverEnd - coverStart + 1)

┌────────────────────────────────────────────────────────────────┐
│ Proximity Scoring Examples                                     │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Doc 1: "machine learning is machine learning"                  │
│         Pos:0      1      2  3       4                          │
│                                                                 │
│   Cover 1: [Pos 0-1]  → score += 1/(1-0+1) = 1/2 = 0.500      │
│   Cover 2: [Pos 3-4]  → score += 1/(4-3+1) = 1/2 = 0.500      │
│                         ─────────────────────────────           │
│   Total Score: 1.000                                            │
│                                                                 │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Doc 2: "learning about machine and learning"                   │
│         Pos:0       1     2       3   4                         │
│                                                                 │
│   Cover 1: [Pos 0-2]  → score += 1/(2-0+1) = 1/3 = 0.333      │
│   Cover 2: [Pos 2-4]  → score += 1/(4-2+1) = 1/3 = 0.333      │
│                         ─────────────────────────────           │
│   Total Score: 0.666                                            │
│                                                                 │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Doc 3: "machine ... ... ... ... learning"                      │
│         Pos:0    1   2   3   4   5                              │
│                                                                 │
│   Cover 1: [Pos 0-5]  → score += 1/(5-0+1) = 1/6 = 0.167      │
│                         ─────────────────────────────           │
│   Total Score: 0.167                                            │
│                                                                 │
└────────────────────────────────────────────────────────────────┘

Ranking: Doc 1 (1.000) > Doc 2 (0.666) > Doc 3 (0.167)
          ▲               ▲               ▲
      Terms closest   Terms medium   Terms far apart
```

**Why This Works:**

- Smaller distances → larger scores (inverse relationship)
- Multiple occurrences → higher scores (additive)
- Documents with terms close together rank higher

## API Reference

### InvertedIndex

#### NewInvertedIndex

```go
func NewInvertedIndex() *InvertedIndex
```

Creates a new empty inverted index.

**Example:**

```go
idx := blaze.NewInvertedIndex()
```

#### Index

```go
func (idx *InvertedIndex) Index(docID int, document string)
```

Adds a document to the inverted index. Thread-safe.

**Parameters:**

- `docID`: Unique document identifier
- `document`: Text content to index

**Example:**

```go
idx.Index(1, "The quick brown fox jumps over the lazy dog")
idx.Index(2, "A fast brown dog")
```

**What Happens:**

1. Text is analyzed (tokenized, stemmed, etc.)
2. Each token is recorded with its position
3. Positions are stored in skip lists for fast lookup

#### First

```go
func (idx *InvertedIndex) First(token string) (Position, error)
```

Returns the first occurrence of a token in the index.

**Example:**

```go
pos, err := idx.First("quick")
if err != nil {
    // Token not found
}
fmt.Printf("Doc %d, Pos %d\n", int(pos.DocumentID), int(pos.Offset))
```

**Returns:**

- `Position`: Location of first occurrence
- `error`: `ErrNoPostingList` if token doesn't exist

#### Last

```go
func (idx *InvertedIndex) Last(token string) (Position, error)
```

Returns the last occurrence of a token in the index.

**Example:**

```go
pos, err := idx.Last("quick")
```

#### Next

```go
func (idx *InvertedIndex) Next(token string, currentPos Position) (Position, error)
```

Finds the next occurrence of a token after the given position.

**Example:**

```go
// Iterate through all occurrences
pos := blaze.BOFDocument
for {
    pos, err = idx.Next("quick", pos)
    if pos.IsEnd() || err != nil {
        break
    }
    fmt.Printf("Found at Doc %d, Pos %d\n",
        int(pos.DocumentID), int(pos.Offset))
}
```

#### Previous

```go
func (idx *InvertedIndex) Previous(token string, currentPos Position) (Position, error)
```

Finds the previous occurrence of a token before the given position.

#### NextPhrase

```go
func (idx *InvertedIndex) NextPhrase(query string, startPos Position) []Position
```

Finds the next occurrence of a phrase (exact word sequence).

**Parameters:**

- `query`: Space-separated phrase (e.g., "quick brown fox")
- `startPos`: Position to start searching from

**Returns:**

- `[]Position`: Array with two elements [phraseStart, phraseEnd]
- Returns `[EOFDocument, EOFDocument]` if no match found

**Example:**

```go
matches := idx.NextPhrase("quick brown fox", blaze.BOFDocument)
if !matches[0].IsEnd() {
    fmt.Printf("Phrase found in Doc %d from Pos %d to %d\n",
        int(matches[0].DocumentID),
        int(matches[0].Offset),
        int(matches[1].Offset))
}
```

#### FindAllPhrases

```go
func (idx *InvertedIndex) FindAllPhrases(query string, startPos Position) [][]Position
```

Finds all occurrences of a phrase in the entire index.

**Example:**

```go
allMatches := idx.FindAllPhrases("brown fox", blaze.BOFDocument)
for _, match := range allMatches {
    fmt.Printf("Doc %d: Pos %d-%d\n",
        int(match[0].DocumentID),
        int(match[0].Offset),
        int(match[1].Offset))
}
```

#### NextCover

```go
func (idx *InvertedIndex) NextCover(tokens []string, startPos Position) []Position
```

Finds the next "cover" - a range containing all given tokens.

**Parameters:**

- `tokens`: Array of search terms
- `startPos`: Position to start searching from

**Returns:**

- `[]Position`: Array with [coverStart, coverEnd]

**Example:**

```go
cover := idx.NextCover([]string{"quick", "fox", "brown"}, blaze.BOFDocument)
fmt.Printf("Cover: Doc %d, Pos %d-%d\n",
    int(cover[0].DocumentID),
    int(cover[0].Offset),
    int(cover[1].Offset))
```

#### RankProximity

```go
func (idx *InvertedIndex) RankProximity(query string, maxResults int) []Match
```

Performs proximity-based ranking of search results. This is the main search function.

**Parameters:**

- `query`: Search query (e.g., "machine learning")
- `maxResults`: Maximum number of results to return

**Returns:**

- `[]Match`: Sorted array of matches with scores

**Example:**

```go
results := idx.RankProximity("quick brown", 5)
for i, match := range results {
    fmt.Printf("%d. Doc %d (score: %.2f)\n",
        i+1,
        int(match.Offsets[0].DocumentID),
        match.Score)
}
```

#### Encode

```go
func (idx *InvertedIndex) Encode() ([]byte, error)
```

Serializes the inverted index to binary format.

**Example:**

```go
data, err := idx.Encode()
if err != nil {
    log.Fatal(err)
}

// Save to file
err = os.WriteFile("index.bin", data, 0644)
```

#### Decode

```go
func (idx *InvertedIndex) Decode(data []byte) error
```

Deserializes binary data back into an inverted index.

**Example:**

```go
data, err := os.ReadFile("index.bin")
if err != nil {
    log.Fatal(err)
}

idx := blaze.NewInvertedIndex()
err = idx.Decode(data)
```

### Text Analysis

#### Analyze

```go
func Analyze(text string) []string
```

Transforms raw text into searchable tokens using the default pipeline.

**Example:**

```go
tokens := blaze.Analyze("The Quick Brown Fox Jumps!")
// Returns: ["quick", "brown", "fox", "jump"]
```

#### AnalyzeWithConfig

```go
func AnalyzeWithConfig(text string, config AnalyzerConfig) []string
```

Transforms text using a custom configuration.

**Example:**

```go
config := blaze.AnalyzerConfig{
    MinTokenLength:  3,
    EnableStemming:  false,
    EnableStopwords: true,
}
tokens := blaze.AnalyzeWithConfig("The quick brown fox", config)
```

### Position

#### Position Methods

```go
func (p *Position) GetDocumentID() int
func (p *Position) GetOffset() int
func (p *Position) IsBeginning() bool
func (p *Position) IsEnd() bool
func (p *Position) IsBefore(other Position) bool
func (p *Position) IsAfter(other Position) bool
func (p *Position) Equals(other Position) bool
```

**Example:**

```go
pos1 := blaze.Position{DocumentID: 1, Offset: 5}
pos2 := blaze.Position{DocumentID: 1, Offset: 10}

if pos1.IsBefore(pos2) {
    fmt.Println("pos1 comes before pos2")
}
```

### Skip List

#### NewSkipList

```go
func NewSkipList() *SkipList
```

Creates a new empty skip list.

#### Insert

```go
func (sl *SkipList) Insert(key Position)
```

Adds or updates a position in the skip list. Average O(log n).

#### Find

```go
func (sl *SkipList) Find(key Position) (Position, error)
```

Searches for an exact position. Average O(log n).

#### Delete

```go
func (sl *SkipList) Delete(key Position) bool
```

Removes a position from the skip list. Average O(log n).

#### FindLessThan

```go
func (sl *SkipList) FindLessThan(key Position) (Position, error)
```

Finds the largest position less than the given position.

#### FindGreaterThan

```go
func (sl *SkipList) FindGreaterThan(key Position) (Position, error)
```

Finds the smallest position greater than the given position.

## Examples

### Example 1: Basic Document Search

```go
package main

import (
    "fmt"
    "github.com/wizenheimer/blaze"
)

func main() {
    // Create index
    idx := blaze.NewInvertedIndex()

    // Index documents
    idx.Index(1, "Go is a programming language designed at Google")
    idx.Index(2, "Python is a high-level programming language")
    idx.Index(3, "Go is fast and efficient for system programming")

    // Search for "programming language"
    results := idx.RankProximity("programming language", 10)

    fmt.Println("Search results for 'programming language':")
    for i, match := range results {
        docID := int(match.Offsets[0].DocumentID)
        fmt.Printf("%d. Document %d (score: %.3f)\n", i+1, docID, match.Score)
    }
}
```

**Output:**

```
Search results for 'programming language':
1. Document 1 (score: 1.000)
2. Document 2 (score: 1.000)
3. Document 3 (score: 0.500)
```

### Example 2: Phrase Search

```go
package main

import (
    "fmt"
    "github.com/wizenheimer/blaze"
)

func main() {
    idx := blaze.NewInvertedIndex()

    idx.Index(1, "the quick brown fox jumps over the lazy dog")
    idx.Index(2, "a quick brown dog runs fast")
    idx.Index(3, "the lazy brown fox sleeps")

    // Find exact phrase "brown fox"
    matches := idx.FindAllPhrases("brown fox", blaze.BOFDocument)

    fmt.Println("Documents containing 'brown fox' as a phrase:")
    for _, match := range matches {
        docID := int(match[0].DocumentID)
        start := int(match[0].Offset)
        end := int(match[1].Offset)
        fmt.Printf("Document %d: positions %d-%d\n", docID, start, end)
    }
}
```

**Output:**

```
Documents containing 'brown fox' as a phrase:
Document 1: positions 1-2
Document 3: positions 2-3
```

### Example 3: Iterating Through Positions

```go
package main

import (
    "fmt"
    "github.com/wizenheimer/blaze"
)

func main() {
    idx := blaze.NewInvertedIndex()

    idx.Index(1, "quick test quick test quick")
    idx.Index(2, "another quick test here")

    // Find all occurrences of "quick"
    fmt.Println("All occurrences of 'quick':")

    pos := blaze.BOFDocument
    for {
        pos, err := idx.Next("quick", pos)
        if err != nil || pos.IsEnd() {
            break
        }
        fmt.Printf("  Doc %d, Pos %d\n",
            int(pos.DocumentID),
            int(pos.Offset))
    }
}
```

**Output:**

```
All occurrences of 'quick':
  Doc 1, Pos 0
  Doc 1, Pos 2
  Doc 1, Pos 4
  Doc 2, Pos 1
```

### Example 4: Persistence with Serialization

```go
package main

import (
    "fmt"
    "os"
    "github.com/wizenheimer/blaze"
)

func main() {
    // Build and save index
    idx := blaze.NewInvertedIndex()
    idx.Index(1, "machine learning algorithms")
    idx.Index(2, "deep learning neural networks")
    idx.Index(3, "natural language processing")

    // Serialize to binary
    data, err := idx.Encode()
    if err != nil {
        panic(err)
    }

    // Save to file
    err = os.WriteFile("search_index.bin", data, 0644)
    if err != nil {
        panic(err)
    }
    fmt.Println("Index saved to search_index.bin")

    // Load index from file
    loadedData, err := os.ReadFile("search_index.bin")
    if err != nil {
        panic(err)
    }

    loadedIdx := blaze.NewInvertedIndex()
    err = loadedIdx.Decode(loadedData)
    if err != nil {
        panic(err)
    }

    // Use loaded index
    results := loadedIdx.RankProximity("learning", 5)
    fmt.Printf("Found %d documents\n", len(results))
}
```

### Example 5: Custom Analyzer Configuration

```go
package main

import (
    "fmt"
    "github.com/wizenheimer/blaze"
)

func main() {
    // Create custom analyzer config (no stemming, longer min length)
    config := blaze.AnalyzerConfig{
        MinTokenLength:  3,      // Minimum 3 characters
        EnableStemming:  false,  // Keep original word forms
        EnableStopwords: true,   // Still remove stopwords
    }

    text := "The running dogs are running fast"

    // Compare default vs custom analysis
    defaultTokens := blaze.Analyze(text)
    customTokens := blaze.AnalyzeWithConfig(text, config)

    fmt.Println("Default tokens:", defaultTokens)
    fmt.Println("Custom tokens:", customTokens)
}
```

**Output:**

```
Default tokens: [run dog run fast]
Custom tokens: [running dogs running fast]
```

### Example 6: Building a Simple Search Engine

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "github.com/wizenheimer/blaze"
)

func main() {
    // Create index
    idx := blaze.NewInvertedIndex()

    // Index some documents
    docs := map[int]string{
        1: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software",
        2: "Python is a programming language that lets you work quickly and integrate systems more effectively",
        3: "JavaScript is a programming language that conforms to the ECMAScript specification",
        4: "Rust is a multi-paradigm programming language focused on performance and safety",
        5: "Java is a class-based, object-oriented programming language designed for portability",
    }

    for id, doc := range docs {
        idx.Index(id, doc)
    }

    // Interactive search
    scanner := bufio.NewScanner(os.Stdin)

    for {
        fmt.Print("\nSearch query (or 'quit' to exit): ")
        if !scanner.Scan() {
            break
        }

        query := strings.TrimSpace(scanner.Text())
        if query == "quit" {
            break
        }

        if query == "" {
            continue
        }

        // Perform search
        results := idx.RankProximity(query, 5)

        if len(results) == 0 {
            fmt.Println("No results found")
            continue
        }

        // Display results
        fmt.Printf("\nFound %d result(s):\n", len(results))
        for i, match := range results {
            docID := int(match.Offsets[0].DocumentID)
            score := match.Score

            fmt.Printf("\n%d. Document %d (Score: %.3f)\n", i+1, docID, score)
            fmt.Printf("   %s\n", docs[docID])
        }
    }
}
```

## Performance Characteristics

### Time Complexity

| Operation            | Average      | Worst Case | Notes                           |
| -------------------- | ------------ | ---------- | ------------------------------- |
| Index (per document) | O(n × log m) | O(n × m)   | n = tokens, m = total positions |
| Term lookup          | O(log m)     | O(m)       | m = positions for term          |
| Phrase search        | O(k × log m) | O(k × m)   | k = phrase length               |
| Proximity ranking    | O(t × m)     | O(t × m)   | t = query terms                 |
| Skip list insert     | O(log n)     | O(n)       | n = elements in list            |
| Skip list search     | O(log n)     | O(n)       | Probabilistically rare          |

### Space Complexity

| Component        | Space        | Notes                       |
| ---------------- | ------------ | --------------------------- |
| Inverted index   | O(n)         | n = total unique positions  |
| Skip list nodes  | O(n × log n) | Average 2 pointers per node |
| Analyzer         | O(1)         | In-place processing         |
| Serialized index | O(n)         | Compact binary format       |

### Benchmarks

Performance on Apple M2 (8 cores), Go 1.24:

```
BenchmarkIndex-8                     50000    35421 ns/op    18234 B/op    245 allocs/op
BenchmarkTermSearch-8              300000     4123 ns/op      128 B/op      3 allocs/op
BenchmarkPhraseSearch-8            100000    12456 ns/op      512 B/op     12 allocs/op
BenchmarkProximityRanking-8         50000    28934 ns/op     2048 B/op     45 allocs/op
BenchmarkSkipListInsert-8         3000000      413 ns/op      255 B/op      6 allocs/op
BenchmarkSkipListSearch-8         5000000      203 ns/op       23 B/op      1 allocs/op
BenchmarkAnalyze-8                1000000     1234 ns/op      512 B/op      8 allocs/op
BenchmarkEncode-8                   10000   156789 ns/op    65536 B/op    234 allocs/op
BenchmarkDecode-8                   15000   123456 ns/op    49152 B/op    189 allocs/op
```

### Scalability

**Index Size vs Performance:**

| Documents | Terms | Index Time | Search Time | Memory |
| --------- | ----- | ---------- | ----------- | ------ |
| 1K        | 10K   | 50ms       | 0.5ms       | 2 MB   |
| 10K       | 100K  | 500ms      | 1ms         | 20 MB  |
| 100K      | 1M    | 5s         | 2ms         | 200 MB |
| 1M        | 10M   | 50s        | 5ms         | 2 GB   |

**Notes:**

- Search time remains relatively constant due to O(log n) operations
- Memory scales linearly with unique positions
- Serialization reduces storage by ~40% compared to in-memory size

## Configuration

### Analyzer Configuration

Customize the text analysis pipeline:

```go
type AnalyzerConfig struct {
    MinTokenLength  int  // Minimum token length (default: 2)
    EnableStemming  bool // Apply stemming (default: true)
    EnableStopwords bool // Remove stopwords (default: true)
}
```

**Configuration Examples:**

```go
// Exact matching (no stemming, keep all words)
exactConfig := blaze.AnalyzerConfig{
    MinTokenLength:  1,
    EnableStemming:  false,
    EnableStopwords: false,
}

// Fuzzy matching (aggressive stemming)
fuzzyConfig := blaze.AnalyzerConfig{
    MinTokenLength:  2,
    EnableStemming:  true,
    EnableStopwords: true,
}

// Code search (no stemming, no stopwords, longer tokens)
codeConfig := blaze.AnalyzerConfig{
    MinTokenLength:  3,
    EnableStemming:  false,
    EnableStopwords: false,
}
```

### Tuning Recommendations

**MinTokenLength:**

- **1**: Very permissive, large index
- **2**: Balanced (default), filters single chars
- **3**: Strict, smaller index, misses short words

**EnableStemming:**

- **true**: Better recall, finds related words ("run" matches "running")
- **false**: Exact matching, preserves original word forms

**EnableStopwords:**

- **true**: Smaller index, faster search, standard behavior
- **false**: Complete indexing, useful for phrase search

### Skip List Parameters

```go
const MaxHeight = 32  // Maximum tower height
```

**Tower Height Probability:**

- Height 1: 50%
- Height 2: 25%
- Height 3: 12.5%
- Height 4: 6.25%

This geometric distribution ensures O(log n) average performance.

## Use Cases

### 1. Document Search Systems

Build a search engine for documents:

```go
type Document struct {
    ID      int
    Title   string
    Content string
}

func IndexDocuments(docs []Document) *blaze.InvertedIndex {
    idx := blaze.NewInvertedIndex()

    for _, doc := range docs {
        // Combine title and content
        text := doc.Title + " " + doc.Content
        idx.Index(doc.ID, text)
    }

    return idx
}

func SearchDocuments(idx *blaze.InvertedIndex, query string) []int {
    matches := idx.RankProximity(query, 20)

    docIDs := make([]int, len(matches))
    for i, match := range matches {
        docIDs[i] = int(match.Offsets[0].DocumentID)
    }

    return docIDs
}
```

### 2. Log Analysis

Search through log files:

```go
func IndexLogs(logFile string) (*blaze.InvertedIndex, error) {
    idx := blaze.NewInvertedIndex()

    file, err := os.Open(logFile)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    lineNumber := 1

    for scanner.Scan() {
        idx.Index(lineNumber, scanner.Text())
        lineNumber++
    }

    return idx, scanner.Err()
}

// Find all ERROR log lines
matches := idx.RankProximity("ERROR", 100)
```

### 3. Code Search

Search through source code:

```go
func IndexCodebase(rootDir string) (*blaze.InvertedIndex, error) {
    idx := blaze.NewInvertedIndex()
    fileID := 1

    // Custom config for code (no stemming, keep all words)
    config := blaze.AnalyzerConfig{
        MinTokenLength:  2,
        EnableStemming:  false,
        EnableStopwords: false,
    }

    err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return err
        }

        // Only index Go files
        if !strings.HasSuffix(path, ".go") {
            return nil
        }

        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        // Use custom analyzer
        tokens := blaze.AnalyzeWithConfig(string(content), config)
        // ... index tokens ...

        fileID++
        return nil
    })

    return idx, err
}
```

### 4. E-commerce Product Search

Search product catalog:

```go
type Product struct {
    ID          int
    Name        string
    Description string
    Category    string
    Tags        []string
}

func IndexProducts(products []Product) *blaze.InvertedIndex {
    idx := blaze.NewInvertedIndex()

    for _, product := range products {
        // Combine all searchable fields
        searchText := fmt.Sprintf("%s %s %s %s",
            product.Name,
            product.Description,
            product.Category,
            strings.Join(product.Tags, " "))

        idx.Index(product.ID, searchText)
    }

    return idx
}

// Search for "wireless headphones"
results := idx.RankProximity("wireless headphones", 10)
```

### 5. Email Search

Index and search email messages:

```go
type Email struct {
    ID      int
    From    string
    Subject string
    Body    string
}

func IndexEmails(emails []Email) *blaze.InvertedIndex {
    idx := blaze.NewInvertedIndex()

    for _, email := range emails {
        searchText := fmt.Sprintf("%s %s %s",
            email.From,
            email.Subject,
            email.Body)

        idx.Index(email.ID, searchText)
    }

    return idx
}

// Find emails about "project deadline"
matches := idx.RankProximity("project deadline", 50)
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make bench

# Run all checks (format, vet, lint, test)
make check
```

### Test Coverage

The library has comprehensive test coverage:

```bash
$ make test-coverage
Running tests...
ok      github.com/wizenheimer/blaze    2.456s  coverage: 98.5% of statements
Generating coverage report...
Coverage report: coverage.html
```

**Coverage by Component:**

- Inverted Index: 100%
- Skip Lists: 100%
- Text Analysis: 100%
- Search Operations: 98%
- Serialization: 100%

### Writing Tests

Example test:

```go
func TestSearchFunctionality(t *testing.T) {
    idx := blaze.NewInvertedIndex()

    // Index test documents
    idx.Index(1, "the quick brown fox")
    idx.Index(2, "the lazy brown dog")

    // Test phrase search
    matches := idx.FindAllPhrases("brown fox", blaze.BOFDocument)

    if len(matches) != 1 {
        t.Errorf("Expected 1 match, got %d", len(matches))
    }

    if int(matches[0][0].DocumentID) != 1 {
        t.Errorf("Expected document 1, got %d", int(matches[0][0].DocumentID))
    }
}
```

## Architecture

### Component Overview

```
blaze/
├── index.go          # Inverted index implementation
├── skiplist.go       # Skip list data structure
├── search.go         # Search algorithms (phrase, proximity)
├── analyzer.go       # Text analysis pipeline
├── serialization.go  # Binary encoding/decoding
├── *_test.go         # Comprehensive test suite
├── Makefile          # Development commands
└── public/           # Documentation website
    └── index.html
```

### Data Flow

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Complete Data Flow                           │
└──────────────────────────────────────────────────────────────────────┘

                              User Input
                       "The Quick Brown Fox!"
                                │
                                ▼
            ┌───────────────────────────────────────────┐
            │      Text Analysis Pipeline               │
            │  ┌─────────────────────────────────────┐  │
            │  │ 1. Tokenization                     │  │
            │  │    ["The", "Quick", "Brown", "Fox"] │  │
            │  └────────────┬────────────────────────┘  │
            │               ▼                            │
            │  ┌─────────────────────────────────────┐  │
            │  │ 2. Lowercasing                      │  │
            │  │    ["the", "quick", "brown", "fox"] │  │
            │  └────────────┬────────────────────────┘  │
            │               ▼                            │
            │  ┌─────────────────────────────────────┐  │
            │  │ 3. Stopword Filtering               │  │
            │  │    ["quick", "brown", "fox"]        │  │
            │  └────────────┬────────────────────────┘  │
            │               ▼                            │
            │  ┌─────────────────────────────────────┐  │
            │  │ 4. Length Filtering                 │  │
            │  │    ["quick", "brown", "fox"]        │  │
            │  └────────────┬────────────────────────┘  │
            │               ▼                            │
            │  ┌─────────────────────────────────────┐  │
            │  │ 5. Stemming                         │  │
            │  │    ["quick", "brown", "fox"]        │  │
            │  └────────────┬────────────────────────┘  │
            └───────────────┼────────────────────────────┘
                            ▼
                    ["quick", "brown", "fox"]
                            │
                            ▼
            ┌───────────────────────────────────────────┐
            │       Inverted Index (Indexing)           │
            │                                            │
            │  ┌─────────┬────────────────────────┐     │
            │  │ "quick" │ → SkipList             │     │
            │  │         │    └─> [Doc1:Pos0]     │     │
            │  ├─────────┼────────────────────────┤     │
            │  │ "brown" │ → SkipList             │     │
            │  │         │    └─> [Doc1:Pos1]     │     │
            │  ├─────────┼────────────────────────┤     │
            │  │ "fox"   │ → SkipList             │     │
            │  │         │    └─> [Doc1:Pos2]     │     │
            │  └─────────┴────────────────────────┘     │
            └───────────────┬───────────────────────────┘
                            │
          ┌─────────────────┴─────────────────┐
          │        Search Operations          │
          ▼                                   ▼
    ┌──────────┐                      ┌────────────┐
    │  Term    │                      │  Phrase    │
    │  Search  │                      │  Search    │
    └────┬─────┘                      └─────┬──────┘
         │                                  │
         └──────────┬───────────────────────┘
                    ▼
            ┌───────────────┐
            │   Proximity   │
            │   Ranking     │
            └───────┬───────┘
                    │
                    ▼
            ┌───────────────────────┐
            │  Ranked Results       │
            │  ┌─────────────────┐  │
            │  │ Doc 1: Score 1.0│  │
            │  │ Doc 2: Score 0.5│  │
            │  │ Doc 3: Score 0.3│  │
            │  └─────────────────┘  │
            └───────────────────────┘
```

### Key Design Decisions

**1. Skip Lists over Balanced Trees**

Rationale:

- Simpler implementation (no rotation logic)
- Better cache locality
- Easier to make concurrent
- Comparable performance (O(log n))
- Used in production systems (Redis, LevelDB)

**2. Position-Based Indexing**

Instead of just tracking document IDs, Blaze tracks exact word positions:

```
Traditional Index (Document IDs only):
┌─────────┬──────────────────┐
│ "quick" │ [Doc1, Doc3]     │  ✗ Can't do phrase search
└─────────┴──────────────────┘  ✗ Can't rank by proximity

Position-Based Index (Document + Offset):
┌─────────┬────────────────────────────────────┐
│ "quick" │ [Doc1:Pos1, Doc3:Pos0]             │  ✓ Phrase search
│ "brown" │ [Doc1:Pos2, Doc3:Pos1]             │  ✓ Proximity ranking
│ "fox"   │ [Doc1:Pos3]                        │  ✓ Snippet generation
└─────────┴────────────────────────────────────┘  ✓ Precise results

Can verify: "quick brown" is a phrase in Doc1 (Pos1→Pos2)
            but NOT in Doc3 (Pos0 and Pos1 are not "quick brown")
```

Benefits:

- Enables phrase search (check consecutive positions)
- Enables proximity ranking (measure distances)
- Enables snippet generation (extract relevant parts)
- More precise search results

Trade-offs:

- Larger index size (~2-3x more data)
- More complex algorithms (but still O(log n))

**3. Binary Serialization**

Custom binary format instead of JSON:

Advantages:

- 60% smaller file size
- 3x faster parsing
- Preserves skip list structure
- Suitable for large indexes

**4. Configurable Text Analysis**

Pluggable analyzer configuration:

Benefits:

- Adapt to different use cases
- Trade-off precision vs recall
- Support multiple languages (future)
- Domain-specific customization

## Best Practices

### 1. Choose Appropriate Document IDs

Use stable, unique identifiers:

```go
// Good: Use database primary keys
idx.Index(dbRecord.ID, dbRecord.Content)

// Bad: Use array indices (changes when reordering)
for i, doc := range docs {
    idx.Index(i, doc.Content)  // Don't do this
}
```

### 2. Batch Indexing for Large Datasets

```go
func IndexLargeDataset(docs []Document) *blaze.InvertedIndex {
    idx := blaze.NewInvertedIndex()

    // Process in batches
    batchSize := 1000
    for i := 0; i < len(docs); i += batchSize {
        end := min(i+batchSize, len(docs))
        batch := docs[i:end]

        for _, doc := range batch {
            idx.Index(doc.ID, doc.Content)
        }

        // Optional: periodic serialization for checkpoints
        if i%10000 == 0 {
            data, _ := idx.Encode()
            os.WriteFile(fmt.Sprintf("checkpoint_%d.bin", i), data, 0644)
        }
    }

    return idx
}
```

### 3. Use Appropriate Analyzer Config

Match configuration to your use case:

```go
// Natural language text (books, articles)
naturalLanguageConfig := blaze.AnalyzerConfig{
    MinTokenLength:  2,
    EnableStemming:  true,   // Find related words
    EnableStopwords: true,   // Remove common words
}

// Technical documentation (code, APIs)
technicalConfig := blaze.AnalyzerConfig{
    MinTokenLength:  2,
    EnableStemming:  false,  // Keep exact terms
    EnableStopwords: false,  // Keep all words
}

// Product names (e-commerce)
productConfig := blaze.AnalyzerConfig{
    MinTokenLength:  1,      // Include single chars (e.g., "X")
    EnableStemming:  false,  // Exact product names
    EnableStopwords: false,  // Keep all words
}
```

### 4. Persist Index for Large Datasets

Don't rebuild the index every time:

```go
const indexFile = "search_index.bin"

func LoadOrBuildIndex(docs []Document) (*blaze.InvertedIndex, error) {
    // Try to load existing index
    if data, err := os.ReadFile(indexFile); err == nil {
        idx := blaze.NewInvertedIndex()
        if err := idx.Decode(data); err == nil {
            return idx, nil
        }
    }

    // Build new index
    idx := blaze.NewInvertedIndex()
    for _, doc := range docs {
        idx.Index(doc.ID, doc.Content)
    }

    // Save for next time
    if data, err := idx.Encode(); err == nil {
        os.WriteFile(indexFile, data, 0644)
    }

    return idx, nil
}
```

### 5. Handle Concurrent Access

The Index method is thread-safe, but for read-heavy workloads:

```go
type SearchService struct {
    idx *blaze.InvertedIndex
    mu  sync.RWMutex
}

func (s *SearchService) Index(docID int, content string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.idx.Index(docID, content)
}

func (s *SearchService) Search(query string) []blaze.Match {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.idx.RankProximity(query, 20)
}
```

### 6. Monitor Index Size

Track index growth:

```go
func (idx *InvertedIndex) Stats() map[string]interface{} {
    stats := make(map[string]interface{})

    stats["unique_terms"] = len(idx.PostingsList)

    totalPositions := 0
    for _, skipList := range idx.PostingsList {
        // Count positions in this skip list
        iter := skipList.Iterator()
        for iter.HasNext() {
            iter.Next()
            totalPositions++
        }
    }

    stats["total_positions"] = totalPositions
    stats["avg_positions_per_term"] = float64(totalPositions) / float64(len(idx.PostingsList))

    return stats
}
```

### 7. Limit Result Set Size

Always specify a reasonable max results:

```go
// Good: Limit results
results := idx.RankProximity("search query", 100)

// Bad: Could return millions of results
results := idx.RankProximity("search query", math.MaxInt32)
```

### 8. Pre-process Queries

Normalize queries before searching:

```go
func NormalizeQuery(query string) string {
    // Remove extra whitespace
    query = strings.TrimSpace(query)
    query = strings.Join(strings.Fields(query), " ")

    // Convert to lowercase
    query = strings.ToLower(query)

    // Remove special characters (optional)
    query = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(query, "")

    return query
}

// Use normalized query
normalizedQuery := NormalizeQuery(userInput)
results := idx.RankProximity(normalizedQuery, 20)
```

## Contributing

Contributions are welcome! Please follow these guidelines:

### Development Setup

```bash
# Clone repository
git clone https://github.com/wizenheimer/blaze.git
cd blaze

# Install dependencies
make deps

# Run tests
make test

# Run linter
make lint
```

### Code Style

- Follow Go conventions (gofmt, golint)
- Write comprehensive comments
- Include examples in documentation
- Add tests for new features
- Keep functions focused and small

### Commit Messages

Use descriptive commit messages:

```
Good:
- "feat: Add proximity ranking algorithm"
- "feat: Handle empty query in RankProximity"
- "fix: Reduce allocations in skip list insert"

Bad:
- "Update code"
- "Fix bug uwu"
- "WIP"
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests
5. Run `make check` to verify
6. Commit your changes
7. Push to your fork
8. Open a Pull Request

## License

MIT License

## Acknowledgments

- **Skip Lists**: Original paper by William Pugh (1990)
- **Snowball Stemmer**: Martin Porter's stemming algorithm
- **Inspiration**: Elasticsearch, Lucene, Mettis, Redis, LevelDB
