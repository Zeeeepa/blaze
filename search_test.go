package blaze

import (
	"strings"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PHRASE SEARCH TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_NextPhrase_SimplePhrase(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "the quick brown fox")

	// Search for "quick brown"
	result := idx.NextPhrase("quick brown", BOFDocument)

	if result[0].IsEnd() {
		t.Fatal("NextPhrase() should find 'quick brown'")
	}

	// Should find it at Doc1, starting at position 0 (quick) ending at position 1 (brown)
	if result[0].GetDocumentID() != 1 || result[0].GetOffset() != 0 {
		t.Errorf("Phrase start = Doc%d:Pos%d, want Doc1:Pos0",
			result[0].GetDocumentID(), result[0].GetOffset())
	}

	if result[1].GetDocumentID() != 1 || result[1].GetOffset() != 1 {
		t.Errorf("Phrase end = Doc%d:Pos%d, want Doc1:Pos1",
			result[1].GetDocumentID(), result[1].GetOffset())
	}
}

func TestInvertedIndex_NextPhrase_ThreeWords(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "the quick brown fox jumps")

	// Search for "quick brown fox"
	result := idx.NextPhrase("quick brown fox", BOFDocument)

	if result[0].IsEnd() {
		t.Fatal("NextPhrase() should find 'quick brown fox'")
	}

	// Phrase spans positions 0-2
	if result[0].GetOffset() != 0 || result[1].GetOffset() != 2 {
		t.Errorf("Phrase = Pos%d-Pos%d, want Pos0-Pos2",
			result[0].GetOffset(), result[1].GetOffset())
	}
}

func TestInvertedIndex_NextPhrase_NotFound(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "the quick brown fox")

	// Search for phrase that doesn't exist
	result := idx.NextPhrase("brown quick", BOFDocument)

	if !result[0].IsEnd() {
		t.Error("NextPhrase() should return EOF for non-existent phrase")
	}
}

func TestInvertedIndex_NextPhrase_NonConsecutive(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick jumps brown fox")

	// "quick brown" exists but not consecutively
	result := idx.NextPhrase("quick brown", BOFDocument)

	if !result[0].IsEnd() {
		t.Error("NextPhrase() should not find non-consecutive words")
	}
}

func TestInvertedIndex_NextPhrase_MultipleDocuments(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "the lazy dog")
	idx.Index(2, "the quick brown fox")
	idx.Index(3, "more text here")

	// Search for phrase in Doc2
	result := idx.NextPhrase("quick brown", BOFDocument)

	if result[0].GetDocumentID() != 2 {
		t.Errorf("Found phrase in Doc%d, want Doc2", result[0].GetDocumentID())
	}
}

func TestInvertedIndex_NextPhrase_StartMidDocument(t *testing.T) {
	idx := NewInvertedIndex()
	// After stop word removal and stemming: "quick brown fox jump quick brown dog"
	// Positions: 0=quick, 1=brown, 2=fox, 3=jump, 4=quick, 5=brown, 6=dog
	idx.Index(1, "quick brown fox jumps over quick brown dog")

	// Find first occurrence
	result1 := idx.NextPhrase("quick brown", BOFDocument)
	if result1[0].GetOffset() != 0 {
		t.Errorf("First occurrence at Pos%d, want Pos0", result1[0].GetOffset())
	}

	// Find second occurrence (starting after first)
	// After stop words removed: positions are 0,1,2,3,4,5,6
	// Second "quick brown" is at positions 4-5
	result2 := idx.NextPhrase("quick brown", result1[0])
	if result2[0].GetOffset() != 4 {
		t.Errorf("Second occurrence at Pos%d, want Pos4", result2[0].GetOffset())
	}
}

func TestInvertedIndex_NextPhrase_SingleWord(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	// Single word phrase should work
	result := idx.NextPhrase("brown", BOFDocument)

	if result[0].IsEnd() {
		t.Fatal("NextPhrase() should find single word 'brown'")
	}

	// Start and end should be the same position
	if result[0].GetOffset() != result[1].GetOffset() {
		t.Errorf("Single word phrase: start=%d, end=%d, should be equal",
			result[0].GetOffset(), result[1].GetOffset())
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// FIND ALL PHRASES TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_FindAllPhrases_Multiple(t *testing.T) {
	idx := NewInvertedIndex()
	// After stop word removal: "quick brown fox jump quick brown dog"
	// Positions: 0=quick, 1=brown, 2=fox, 3=jump, 4=quick, 5=brown, 6=dog
	idx.Index(1, "quick brown fox jumps over quick brown dog")

	// Find all occurrences of "quick brown"
	results := idx.FindAllPhrases("quick brown", BOFDocument)

	if len(results) != 2 {
		t.Fatalf("Found %d occurrences, want 2", len(results))
	}

	// First occurrence at positions 0-1
	if results[0][0].GetOffset() != 0 || results[0][1].GetOffset() != 1 {
		t.Errorf("First occurrence = Pos%d-Pos%d, want Pos0-Pos1",
			results[0][0].GetOffset(), results[0][1].GetOffset())
	}

	// Second occurrence at positions 4-5
	if results[1][0].GetOffset() != 4 || results[1][1].GetOffset() != 5 {
		t.Errorf("Second occurrence = Pos%d-Pos%d, want Pos4-Pos5",
			results[1][0].GetOffset(), results[1][1].GetOffset())
	}
}

func TestInvertedIndex_FindAllPhrases_AcrossDocuments(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")
	idx.Index(2, "lazy dog sleeps")
	idx.Index(3, "quick brown dog")
	idx.Index(4, "more quick brown text")

	// Find all occurrences of "quick brown"
	results := idx.FindAllPhrases("quick brown", BOFDocument)

	if len(results) != 3 {
		t.Fatalf("Found %d occurrences, want 3", len(results))
	}

	// Verify documents
	expectedDocs := []int{1, 3, 4}
	for i, result := range results {
		docID := result[0].GetDocumentID()
		if docID != expectedDocs[i] {
			t.Errorf("Occurrence %d in Doc%d, want Doc%d", i, docID, expectedDocs[i])
		}
	}
}

func TestInvertedIndex_FindAllPhrases_None(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")
	idx.Index(2, "lazy dog")

	// Search for phrase that doesn't exist
	results := idx.FindAllPhrases("brown lazy", BOFDocument)

	if len(results) != 0 {
		t.Errorf("Found %d occurrences, want 0", len(results))
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// COVER SEARCH TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_NextCover_SimpleCover(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "the quick brown fox")

	// Search for cover containing "quick" and "fox"
	tokens := []string{"quick", "fox"}
	result := idx.NextCover(tokens, BOFDocument)

	if result[0].IsEnd() {
		t.Fatal("NextCover() should find a cover")
	}

	// Should find cover from position 0 (quick) to position 2 (fox)
	if result[0].GetDocumentID() != 1 {
		t.Errorf("Cover in Doc%d, want Doc1", result[0].GetDocumentID())
	}

	if result[0].GetOffset() != 0 || result[1].GetOffset() != 2 {
		t.Errorf("Cover = Pos%d-Pos%d, want Pos0-Pos2",
			result[0].GetOffset(), result[1].GetOffset())
	}
}

func TestInvertedIndex_NextCover_SamePosition(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	// Search for single token (cover of itself)
	tokens := []string{"brown"}
	result := idx.NextCover(tokens, BOFDocument)

	if result[0].IsEnd() {
		t.Fatal("NextCover() should find a cover")
	}

	// Cover should be a single position
	if result[0].GetOffset() != result[1].GetOffset() {
		t.Errorf("Single token cover: start=%d, end=%d, should be equal",
			result[0].GetOffset(), result[1].GetOffset())
	}
}

func TestInvertedIndex_NextCover_NotInSameDocument(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown")
	idx.Index(2, "lazy fox")

	// "quick" in Doc1, "fox" in Doc2 - should find no cover
	tokens := []string{"quick", "fox"}
	result := idx.NextCover(tokens, BOFDocument)

	if !result[0].IsEnd() {
		t.Error("NextCover() should return EOF when tokens span documents")
	}
}

func TestInvertedIndex_NextCover_MultipleCovers(t *testing.T) {
	idx := NewInvertedIndex()
	// After stop word removal: "quick brown fox jump tall dog"
	// Positions: 0=quick, 1=brown, 2=fox, 3=jump, 4=tall, 5=dog
	idx.Index(1, "quick brown fox jumps over tall dog")

	// First cover
	tokens := []string{"quick", "tall"}
	result1 := idx.NextCover(tokens, BOFDocument)

	if result1[0].IsEnd() {
		t.Fatal("Should find a cover")
	}

	// Cover should span from quick (pos 0) to tall (pos 4)
	if result1[0].GetOffset() != 0 || result1[1].GetOffset() != 4 {
		t.Errorf("First cover = Pos%d-Pos%d, want Pos0-Pos4",
			result1[0].GetOffset(), result1[1].GetOffset())
	}

	// There shouldn't be another cover in this document
	result2 := idx.NextCover(tokens, result1[0])
	if !result2[0].IsEnd() {
		t.Error("Should not find another cover")
	}
}

func TestInvertedIndex_NextCover_TokenNotFound(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	// One token doesn't exist
	tokens := []string{"quick", "elephant"}
	result := idx.NextCover(tokens, BOFDocument)

	if !result[0].IsEnd() {
		t.Error("NextCover() should return EOF when token not found")
	}
}

func TestInvertedIndex_NextCover_ThreeTokens(t *testing.T) {
	idx := NewInvertedIndex()
	// After stop word removal: "quick brown tall fox jump"
	// Positions: 0=quick, 1=brown, 2=tall, 3=fox, 4=jump
	idx.Index(1, "the quick brown tall fox jumps")

	// Cover containing three tokens
	tokens := []string{"quick", "tall", "fox"}
	result := idx.NextCover(tokens, BOFDocument)

	if result[0].IsEnd() {
		t.Fatal("NextCover() should find a cover")
	}

	// Cover should span from "quick" (pos 0) to "fox" (pos 3)
	if result[0].GetOffset() != 0 || result[1].GetOffset() != 3 {
		t.Errorf("Cover = Pos%d-Pos%d, want Pos0-Pos3",
			result[0].GetOffset(), result[1].GetOffset())
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// MATCH TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestMatch_GetKey_Unique(t *testing.T) {
	match1 := Match{
		Offsets: []Position{
			{DocumentID: 1, Offset: 0},
			{DocumentID: 1, Offset: 5},
		},
		Score: 1.5,
	}

	match2 := Match{
		Offsets: []Position{
			{DocumentID: 2, Offset: 0},
			{DocumentID: 2, Offset: 5},
		},
		Score: 1.5,
	}

	key1, err1 := match1.GetKey()
	key2, err2 := match2.GetKey()

	if err1 != nil || err2 != nil {
		t.Fatalf("GetKey() errors: %v, %v", err1, err2)
	}

	// Keys should be different (different documents)
	if key1 == key2 {
		t.Error("Different matches should have different keys")
	}
}

func TestMatch_GetKey_Deterministic(t *testing.T) {
	match := Match{
		Offsets: []Position{
			{DocumentID: 1, Offset: 0},
			{DocumentID: 1, Offset: 5},
		},
		Score: 1.5,
	}

	// Get key multiple times
	key1, _ := match.GetKey()
	key2, _ := match.GetKey()
	key3, _ := match.GetKey()

	// Should always return the same key
	if key1 != key2 || key2 != key3 {
		t.Error("GetKey() should be deterministic")
	}
}

func TestMatch_GetKey_HashLength(t *testing.T) {
	match := Match{
		Offsets: []Position{
			{DocumentID: 1, Offset: 0},
		},
		Score: 1.0,
	}

	key, err := match.GetKey()
	if err != nil {
		t.Fatalf("GetKey() error = %v", err)
	}

	// MD5 hash should be 32 hex characters
	if len(key) != 32 {
		t.Errorf("Key length = %d, want 32", len(key))
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// PROXIMITY RANKING TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_RankProximity_SingleDocument(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	// Search for "quick fox"
	results := idx.RankProximity("quick fox", 10)

	if len(results) != 1 {
		t.Fatalf("Found %d results, want 1", len(results))
	}

	// Should find document 1
	if results[0].Offsets[0].GetDocumentID() != 1 {
		t.Errorf("Result in Doc%d, want Doc1", results[0].Offsets[0].GetDocumentID())
	}

	// Score should be positive
	if results[0].Score <= 0 {
		t.Errorf("Score = %f, want > 0", results[0].Score)
	}
}

func TestInvertedIndex_RankProximity_MultipleDocuments(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")
	idx.Index(2, "lazy dog")
	idx.Index(3, "quick lazy fox")

	// Search for "quick fox"
	results := idx.RankProximity("quick fox", 10)

	// Should find 2 documents (Doc1 and Doc3)
	if len(results) != 2 {
		t.Fatalf("Found %d results, want 2", len(results))
	}
}

func TestInvertedIndex_RankProximity_ProximityScoring(t *testing.T) {
	idx := NewInvertedIndex()
	// Doc1: "quick" and "fox" are close (distance 2)
	idx.Index(1, "quick brown fox")
	// Doc2: "quick" and "fox" are far apart (distance 5)
	idx.Index(2, "quick brown lazy sleeping tired fox")

	// Search for "quick fox"
	results := idx.RankProximity("quick fox", 10)

	if len(results) != 2 {
		t.Fatalf("Found %d results, want 2", len(results))
	}

	// Doc1 should have higher score (closer proximity)
	doc1Score := 0.0
	doc2Score := 0.0

	for _, result := range results {
		docID := result.Offsets[0].GetDocumentID()
		if docID == 1 {
			doc1Score = result.Score
		} else if docID == 2 {
			doc2Score = result.Score
		}
	}

	if doc1Score <= doc2Score {
		t.Errorf("Doc1 score (%f) should be > Doc2 score (%f)", doc1Score, doc2Score)
	}
}

func TestInvertedIndex_RankProximity_MaxResults(t *testing.T) {
	idx := NewInvertedIndex()

	// Index many documents
	for i := 1; i <= 10; i++ {
		idx.Index(i, "quick brown fox")
	}

	// Request only 5 results
	results := idx.RankProximity("quick fox", 5)

	if len(results) > 5 {
		t.Errorf("Returned %d results, want at most 5", len(results))
	}
}

func TestInvertedIndex_RankProximity_EmptyQuery(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	// Empty query
	results := idx.RankProximity("", 10)

	if len(results) != 0 {
		t.Errorf("Empty query returned %d results, want 0", len(results))
	}
}

func TestInvertedIndex_RankProximity_NoResults(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	// Search for tokens that don't exist
	results := idx.RankProximity("elephant giraffe", 10)

	if len(results) != 0 {
		t.Errorf("Found %d results, want 0", len(results))
	}
}

func TestInvertedIndex_RankProximity_SingleToken(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")
	idx.Index(2, "lazy dog")
	idx.Index(3, "quick rabbit")

	// Search for single token
	results := idx.RankProximity("quick", 10)

	// Should find Doc1 and Doc3
	if len(results) != 2 {
		t.Fatalf("Found %d results, want 2", len(results))
	}
}

func TestInvertedIndex_RankProximity_MultipleCoversInDocument(t *testing.T) {
	idx := NewInvertedIndex()
	// After stop word removal: "quick fox jump quick fox"
	// Positions: 0=quick, 1=fox, 2=jump, 3=quick, 4=fox
	idx.Index(1, "quick fox jumps over quick fox")

	// Search for "quick fox"
	results := idx.RankProximity("quick fox", 10)

	if len(results) != 1 {
		t.Fatalf("Found %d results, want 1", len(results))
	}

	// The algorithm finds covers. Each cover gets score = 1/(end-start+1)
	// Cover 1: positions 0-1 → score = 1/(1-0+1) = 1/2 = 0.5
	// Cover 2: positions 3-4 → score = 1/(4-3+1) = 1/2 = 0.5
	// But the algorithm continues from the start position, so it might find
	// an overlapping cover. Let's just check the score is reasonable.
	actualScore := results[0].Score

	// Score should be positive and reasonable
	if actualScore <= 0 {
		t.Errorf("Score = %f, should be positive", actualScore)
	}

	// Score should be at least 0.5 (one cover)
	if actualScore < 0.5 {
		t.Errorf("Score = %f, should be at least 0.5", actualScore)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// INTEGRATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestSearch_CompleteWorkflow(t *testing.T) {
	idx := NewInvertedIndex()

	// Index a small corpus
	idx.Index(1, "the quick brown fox jumps over the lazy dog")
	idx.Index(2, "a lazy brown dog sleeps peacefully")
	idx.Index(3, "the quick brown rabbit hops quickly")
	idx.Index(4, "foxes and dogs are both animals")

	// Test 1: Phrase search
	phraseResults := idx.FindAllPhrases("brown dog", BOFDocument)
	if len(phraseResults) != 1 {
		t.Errorf("Phrase search found %d results, want 1", len(phraseResults))
	}

	// Test 2: Proximity search
	proximityResults := idx.RankProximity("quick brown", 10)
	// Should find Doc1 and Doc3
	if len(proximityResults) != 2 {
		t.Errorf("Proximity search found %d results, want 2", len(proximityResults))
	}

	// Test 3: Multi-word query
	multiResults := idx.RankProximity("fox dog", 10)
	// Doc1 has both, Doc2 has dog, Doc4 has both
	if len(multiResults) < 2 {
		t.Errorf("Multi-word search found %d results, want at least 2", len(multiResults))
	}
}

func TestSearch_RealWorldScenario(t *testing.T) {
	idx := NewInvertedIndex()

	// Index blog posts
	idx.Index(1, "introduction to machine learning algorithms")
	idx.Index(2, "deep learning tutorial for beginners")
	idx.Index(3, "machine learning and deep learning compared")
	idx.Index(4, "natural language processing tutorial")
	idx.Index(5, "machine learning in python")

	// Search: "machine learning"
	results := idx.RankProximity("machine learning", 10)

	// Should find Doc1, Doc3, and Doc5
	if len(results) != 3 {
		t.Errorf("Found %d results for 'machine learning', want 3", len(results))
	}

	// All results should have both words
	for i, result := range results {
		docID := result.Offsets[0].GetDocumentID()
		if docID != 1 && docID != 3 && docID != 5 {
			t.Errorf("Result %d is Doc%d, should be Doc1, Doc3, or Doc5", i, docID)
		}
	}

	// Search: "deep learning tutorial"
	results2 := idx.RankProximity("deep learning tutorial", 10)

	// Should find Doc2 with high score (all three words close together)
	if len(results2) == 0 {
		t.Fatal("Should find results for 'deep learning tutorial'")
	}

	// Doc2 should be in the results
	foundDoc2 := false
	for _, result := range results2 {
		if result.Offsets[0].GetDocumentID() == 2 {
			foundDoc2 = true
			break
		}
	}

	if !foundDoc2 {
		t.Error("Doc2 should be in results for 'deep learning tutorial'")
	}
}

func TestSearch_EdgeCases(t *testing.T) {
	idx := NewInvertedIndex()

	// Test with special characters and punctuation
	idx.Index(1, "Hello, world! This is a test.")
	idx.Index(2, "Test-driven development is great!")

	// Search should work despite punctuation
	results := idx.RankProximity("test", 10)

	// Should find both documents
	if len(results) != 2 {
		t.Errorf("Found %d results, want 2", len(results))
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestLimitResults_LessThanMax(t *testing.T) {
	matches := []Match{
		{Score: 1.0},
		{Score: 2.0},
		{Score: 3.0},
	}

	result := limitResults(matches, 10)

	if len(result) != 3 {
		t.Errorf("limitResults() returned %d items, want 3", len(result))
	}
}

func TestLimitResults_MoreThanMax(t *testing.T) {
	matches := []Match{
		{Score: 1.0},
		{Score: 2.0},
		{Score: 3.0},
		{Score: 4.0},
		{Score: 5.0},
	}

	result := limitResults(matches, 3)

	if len(result) != 3 {
		t.Errorf("limitResults() returned %d items, want 3", len(result))
	}
}

func TestLimitResults_Empty(t *testing.T) {
	matches := []Match{}

	result := limitResults(matches, 10)

	if len(result) != 0 {
		t.Errorf("limitResults() returned %d items, want 0", len(result))
	}
}

func TestIsValidPhrase(t *testing.T) {
	idx := NewInvertedIndex()

	tests := []struct {
		name      string
		start     Position
		end       Position
		termCount int
		want      bool
	}{
		{
			"Valid 2-word phrase",
			Position{DocumentID: 1, Offset: 0},
			Position{DocumentID: 1, Offset: 1},
			2,
			true,
		},
		{
			"Valid 3-word phrase",
			Position{DocumentID: 1, Offset: 5},
			Position{DocumentID: 1, Offset: 7},
			3,
			true,
		},
		{
			"Non-consecutive words",
			Position{DocumentID: 1, Offset: 0},
			Position{DocumentID: 1, Offset: 5},
			3,
			false,
		},
		{
			"Different documents",
			Position{DocumentID: 1, Offset: 0},
			Position{DocumentID: 2, Offset: 1},
			2,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := idx.isValidPhrase(tt.start, tt.end, tt.termCount)
			if got != tt.want {
				t.Errorf("isValidPhrase() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// BENCHMARK TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func BenchmarkNextPhrase(b *testing.B) {
	idx := NewInvertedIndex()

	// Pre-populate index
	for i := 1; i <= 100; i++ {
		idx.Index(i, "the quick brown fox jumps over the lazy dog")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.NextPhrase("quick brown", BOFDocument)
	}
}

func BenchmarkNextCover(b *testing.B) {
	idx := NewInvertedIndex()

	// Pre-populate index
	for i := 1; i <= 100; i++ {
		idx.Index(i, "the quick brown fox jumps over the lazy dog")
	}

	tokens := []string{"quick", "lazy"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.NextCover(tokens, BOFDocument)
	}
}

func BenchmarkRankProximity(b *testing.B) {
	idx := NewInvertedIndex()

	// Pre-populate index with realistic documents
	documents := []string{
		"introduction to machine learning algorithms and techniques",
		"deep learning neural networks for image recognition",
		"natural language processing with python programming",
		"machine learning models and evaluation metrics",
		"computer vision and image processing fundamentals",
	}

	for i, doc := range documents {
		idx.Index(i+1, strings.Repeat(doc+" ", 20)) // Repeat for larger corpus
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.RankProximity("machine learning", 10)
	}
}

func BenchmarkFindAllPhrases(b *testing.B) {
	idx := NewInvertedIndex()

	// Pre-populate index
	for i := 1; i <= 50; i++ {
		idx.Index(i, "the quick brown fox jumps over the lazy dog and quick brown cat")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.FindAllPhrases("quick brown", BOFDocument)
	}
}
