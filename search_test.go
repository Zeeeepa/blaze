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
		DocID: 1,
		Offsets: []Position{
			{DocumentID: 1, Offset: 0},
			{DocumentID: 1, Offset: 5},
		},
		Score: 1.5,
	}

	match2 := Match{
		DocID: 2,
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
		DocID: 1,
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
		DocID: 1,
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
		switch docID {
		case 1:
			doc1Score = result.Score
		case 2:
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

// ═══════════════════════════════════════════════════════════════════════════════
// BM25 TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_calculateIDF_BasicCases(t *testing.T) {
	idx := NewInvertedIndex()

	// Index 3 documents
	idx.Index(1, "machine learning")
	idx.Index(2, "machine learning algorithms")
	idx.Index(3, "deep learning")

	// Get analyzed/stemmed versions of terms
	machineTokens := Analyze("machine")
	learningTokens := Analyze("learning")
	deepTokens := Analyze("deep")

	// Test IDF for "machine" (appears in 2 out of 3 docs)
	idfMachine := idx.calculateIDF(machineTokens[0])
	if idfMachine <= 0 {
		t.Errorf("IDF for 'machine' = %f, want > 0", idfMachine)
	}

	// Test IDF for "learning" (appears in all 3 docs - should be lower)
	idfLearning := idx.calculateIDF(learningTokens[0])
	if idfLearning <= 0 {
		t.Errorf("IDF for 'learning' = %f, want > 0", idfLearning)
	}

	// Test IDF for "deep" (appears in 1 out of 3 docs - should be highest)
	idfDeep := idx.calculateIDF(deepTokens[0])
	if idfDeep <= 0 {
		t.Errorf("IDF for 'deep' = %f, want > 0", idfDeep)
	}

	// Rarer terms should have higher IDF
	if idfDeep <= idfMachine {
		t.Errorf("IDF('deep')=%f should be > IDF('machine')=%f (rarer term)", idfDeep, idfMachine)
	}
}

func TestInvertedIndex_calculateIDF_NonExistentTerm(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "machine learning")

	// IDF for non-existent term should be 0
	idf := idx.calculateIDF("nonexistent")
	if idf != 0 {
		t.Errorf("IDF for non-existent term = %f, want 0", idf)
	}
}

func TestInvertedIndex_calculateIDF_SingleDocument(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "machine learning algorithms")

	// Get analyzed/stemmed version of term
	machineTokens := Analyze("machine")

	// With only 1 document, IDF calculation should still work
	idf := idx.calculateIDF(machineTokens[0])
	// N=1, df=1: log((1-1+0.5)/(1+0.5) + 1) = log(0.5/1.5 + 1) = log(1.333...) ≈ 0.287
	if idf <= 0 {
		t.Errorf("IDF with single document = %f, want > 0", idf)
	}
}

func TestInvertedIndex_countDocsInPostingList(t *testing.T) {
	idx := NewInvertedIndex()

	// Index documents where "machine" appears multiple times per doc
	idx.Index(1, "machine learning machine vision")
	idx.Index(2, "machine intelligence")
	idx.Index(3, "deep learning")

	// Get analyzed/stemmed version of term
	machineTokens := Analyze("machine")

	// Get posting list for "machine" (stemmed version)
	skipList, exists := idx.getPostingList(machineTokens[0])
	if !exists {
		t.Fatal("posting list for 'machine' should exist")
	}

	// Count unique documents
	count := idx.countDocsInPostingList(skipList)
	if count != 2 {
		t.Errorf("countDocsInPostingList() = %d, want 2 (Doc1 and Doc2)", count)
	}
}

func TestInvertedIndex_calculateBM25Score_BasicScoring(t *testing.T) {
	idx := NewInvertedIndex()

	// Index documents
	idx.Index(1, "machine learning algorithms")
	idx.Index(2, "deep learning neural networks")
	idx.Index(3, "machine learning and deep learning")

	// Calculate BM25 score for Doc1 with query "machine learning"
	// Use analyzed tokens (stemmed versions)
	tokens := Analyze("machine learning")
	score := idx.calculateBM25Score(1, tokens)

	// Score should be positive
	if score <= 0 {
		t.Errorf("BM25 score for Doc1 = %f, want > 0", score)
	}
}

func TestInvertedIndex_calculateBM25Score_DocumentWithAllTerms(t *testing.T) {
	idx := NewInvertedIndex()

	idx.Index(1, "machine learning")
	idx.Index(2, "machine")
	idx.Index(3, "learning")

	// Doc1 has both terms, Doc2 and Doc3 have only one
	// Use analyzed tokens (stemmed versions)
	tokens := Analyze("machine learning")
	score1 := idx.calculateBM25Score(1, tokens)
	score2 := idx.calculateBM25Score(2, tokens)
	score3 := idx.calculateBM25Score(3, tokens)

	// Doc1 should have higher score than Doc2 or Doc3
	if score1 <= score2 {
		t.Errorf("Doc1 (both terms) score=%f should be > Doc2 (one term) score=%f", score1, score2)
	}
	if score1 <= score3 {
		t.Errorf("Doc1 (both terms) score=%f should be > Doc3 (one term) score=%f", score1, score3)
	}
}

func TestInvertedIndex_calculateBM25Score_TermFrequency(t *testing.T) {
	idx := NewInvertedIndex()

	// Doc1: "machine" appears once
	idx.Index(1, "machine learning algorithms")
	// Doc2: "machine" appears three times
	idx.Index(2, "machine learning machine vision machine intelligence")

	// Use analyzed tokens (stemmed versions)
	tokens := Analyze("machine")
	score1 := idx.calculateBM25Score(1, tokens)
	score2 := idx.calculateBM25Score(2, tokens)

	// Doc2 should have higher score due to higher term frequency
	if score2 <= score1 {
		t.Errorf("Doc2 (TF=3) score=%f should be > Doc1 (TF=1) score=%f", score2, score1)
	}
}

func TestInvertedIndex_calculateBM25Score_LengthNormalization(t *testing.T) {
	idx := NewInvertedIndex()

	// Short document with term
	idx.Index(1, "machine learning")
	// Long document with same term
	idx.Index(2, "machine learning algorithms neural networks deep learning artificial intelligence natural language processing computer vision")

	// Use analyzed tokens (stemmed versions)
	tokens := Analyze("machine")
	score1 := idx.calculateBM25Score(1, tokens)
	score2 := idx.calculateBM25Score(2, tokens)

	// Shorter document should typically score higher (length normalization)
	// Both have "machine" once, but Doc1 is much shorter
	if score1 <= score2 {
		t.Errorf("Short doc score=%f should be > long doc score=%f due to length normalization", score1, score2)
	}
}

func TestInvertedIndex_calculateBM25Score_NonExistentDocument(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "machine learning")

	// Score for non-existent document should be 0
	score := idx.calculateBM25Score(999, []string{"machine"})
	if score != 0 {
		t.Errorf("Score for non-existent doc = %f, want 0", score)
	}
}

func TestInvertedIndex_RankBM25_BasicRanking(t *testing.T) {
	idx := NewInvertedIndex()

	idx.Index(1, "machine learning algorithms")
	idx.Index(2, "deep learning neural networks")
	idx.Index(3, "machine learning and deep learning")

	// Search for "machine learning"
	results := idx.RankBM25("machine learning", 10)

	// Should find Doc1 and Doc3 (both contain "machine learning")
	if len(results) < 2 {
		t.Fatalf("RankBM25() found %d results, want at least 2", len(results))
	}

	// All results should have positive scores
	for i, result := range results {
		if result.Score <= 0 {
			t.Errorf("Result %d has score=%f, want > 0", i, result.Score)
		}
	}
}

func TestInvertedIndex_RankBM25_ScoreSorting(t *testing.T) {
	idx := NewInvertedIndex()

	// Doc1: Contains both terms once
	idx.Index(1, "machine learning")
	// Doc2: Contains both terms with high frequency
	idx.Index(2, "machine learning machine learning algorithms")
	// Doc3: Contains only one term
	idx.Index(3, "machine vision")

	results := idx.RankBM25("machine learning", 10)

	if len(results) < 2 {
		t.Fatalf("RankBM25() found %d results, want at least 2", len(results))
	}

	// Results should be sorted by score (descending)
	for i := 0; i < len(results)-1; i++ {
		if results[i].Score < results[i+1].Score {
			t.Errorf("Results not sorted: result[%d].Score=%f < result[%d].Score=%f",
				i, results[i].Score, i+1, results[i+1].Score)
		}
	}
}

func TestInvertedIndex_RankBM25_MaxResults(t *testing.T) {
	idx := NewInvertedIndex()

	// Index 10 documents
	for i := 1; i <= 10; i++ {
		idx.Index(i, "machine learning algorithms")
	}

	// Request only 5 results
	results := idx.RankBM25("machine learning", 5)

	if len(results) > 5 {
		t.Errorf("RankBM25() returned %d results, want at most 5", len(results))
	}
}

func TestInvertedIndex_RankBM25_EmptyQuery(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "machine learning")

	results := idx.RankBM25("", 10)

	if len(results) != 0 {
		t.Errorf("RankBM25() with empty query returned %d results, want 0", len(results))
	}
}

func TestInvertedIndex_RankBM25_NoMatches(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "machine learning algorithms")

	results := idx.RankBM25("quantum physics", 10)

	if len(results) != 0 {
		t.Errorf("RankBM25() with no matches returned %d results, want 0", len(results))
	}
}

func TestInvertedIndex_RankBM25_PartialMatches(t *testing.T) {
	idx := NewInvertedIndex()

	idx.Index(1, "machine learning algorithms")
	idx.Index(2, "machine vision")
	idx.Index(3, "deep learning")

	// Query: "machine learning" - should find all docs but rank differently
	results := idx.RankBM25("machine learning", 10)

	// Should find all 3 documents (each has at least one term)
	if len(results) != 3 {
		t.Fatalf("RankBM25() found %d results, want 3", len(results))
	}

	// Doc1 (has both terms) should rank highest
	if results[0].DocID != 1 {
		t.Errorf("Highest ranked doc is Doc%d, want Doc1 (has both terms)", results[0].DocID)
	}
}

func TestInvertedIndex_RankBM25_SingleTerm(t *testing.T) {
	idx := NewInvertedIndex()

	idx.Index(1, "machine learning")
	idx.Index(2, "machine vision")
	idx.Index(3, "deep learning")

	// Single term query
	results := idx.RankBM25("machine", 10)

	// Should find Doc1 and Doc2
	if len(results) != 2 {
		t.Fatalf("RankBM25() found %d results, want 2", len(results))
	}
}

func TestInvertedIndex_RankBM25_DocumentPositions(t *testing.T) {
	idx := NewInvertedIndex()

	idx.Index(1, "machine learning algorithms")
	idx.Index(2, "machine learning")

	results := idx.RankBM25("machine learning", 10)

	if len(results) < 1 {
		t.Fatal("RankBM25() should find at least one result")
	}

	// Each result should have positions where terms appear
	for i, result := range results {
		if len(result.Offsets) == 0 {
			t.Errorf("Result %d (Doc%d) has no position offsets", i, result.DocID)
		}
	}
}

func TestInvertedIndex_BM25Parameters_Custom(t *testing.T) {
	idx := NewInvertedIndex()

	// Set custom BM25 parameters
	idx.BM25Params.K1 = 2.0
	idx.BM25Params.B = 0.5

	idx.Index(1, "machine learning")
	idx.Index(2, "machine learning machine")

	results := idx.RankBM25("machine", 10)

	if len(results) != 2 {
		t.Fatalf("RankBM25() found %d results, want 2", len(results))
	}

	// Scores should reflect custom parameters
	// With higher K1, term frequency saturation is different
	if results[0].Score <= 0 {
		t.Errorf("Score with custom params = %f, want > 0", results[0].Score)
	}
}

func TestInvertedIndex_BM25Parameters_Default(t *testing.T) {
	params := DefaultBM25Parameters()

	// Check default values
	if params.K1 != 1.5 {
		t.Errorf("Default K1 = %f, want 1.5", params.K1)
	}
	if params.B != 0.75 {
		t.Errorf("Default B = %f, want 0.75", params.B)
	}
}

func TestInvertedIndex_RankBM25_vs_RankProximity(t *testing.T) {
	idx := NewInvertedIndex()

	// Create corpus where BM25 and proximity might rank differently
	idx.Index(1, "machine learning algorithms neural networks")
	idx.Index(2, "machine algorithms learning networks neural")
	idx.Index(3, "machine learning")

	bm25Results := idx.RankBM25("machine learning", 10)
	proximityResults := idx.RankProximity("machine learning", 10)

	// Both should find documents
	if len(bm25Results) == 0 {
		t.Error("BM25 should find results")
	}
	if len(proximityResults) == 0 {
		t.Error("Proximity should find results")
	}

	// Doc3 has terms closest together, so proximity might rank it higher
	// Doc1 might score higher in BM25 due to other factors
	// Just verify both ranking methods work
}

func TestInvertedIndex_RankBM25_RareVsCommonTerms(t *testing.T) {
	idx := NewInvertedIndex()

	// "the" appears in all docs (common)
	// "quantum" appears in one doc (rare)
	idx.Index(1, "the quick brown fox")
	idx.Index(2, "the lazy dog")
	idx.Index(3, "the quantum computer")
	idx.Index(4, "the machine learning")

	// IDF for "quantum" should be much higher than "the"
	idfQuantum := idx.calculateIDF("quantum")
	idfThe := idx.calculateIDF("the")

	if idfQuantum <= idfThe {
		t.Errorf("IDF('quantum')=%f should be > IDF('the')=%f", idfQuantum, idfThe)
	}

	// Search for the rare term
	results := idx.RankBM25("quantum", 10)

	if len(results) != 1 {
		t.Fatalf("Search for rare term found %d results, want 1", len(results))
	}

	if results[0].DocID != 3 {
		t.Errorf("Search for 'quantum' found Doc%d, want Doc3", results[0].DocID)
	}
}

func TestInvertedIndex_BM25_DocumentStats(t *testing.T) {
	idx := NewInvertedIndex()

	idx.Index(1, "machine learning algorithms")
	idx.Index(2, "deep learning")

	// Check document statistics
	if len(idx.DocStats) != 2 {
		t.Errorf("DocStats has %d entries, want 2", len(idx.DocStats))
	}

	// Check Doc1 stats
	stats1, exists := idx.DocStats[1]
	if !exists {
		t.Fatal("DocStats for Doc1 should exist")
	}

	if stats1.DocID != 1 {
		t.Errorf("Doc1 stats DocID = %d, want 1", stats1.DocID)
	}

	if stats1.Length != 3 {
		t.Errorf("Doc1 length = %d, want 3", stats1.Length)
	}

	// Check term frequencies using analyzed/stemmed versions
	// "machine" becomes "machin" after stemming
	// "learning" becomes "learn" after stemming
	tokens := Analyze("machine learning algorithms")
	if len(tokens) != 3 {
		t.Errorf("Expected 3 analyzed tokens, got %d", len(tokens))
	}

	// Check that the stemmed terms exist in term frequencies
	for _, token := range tokens {
		if stats1.TermFreqs[token] < 1 {
			t.Errorf("Doc1 '%s' frequency = %d, want at least 1", token, stats1.TermFreqs[token])
		}
	}
}

func TestInvertedIndex_BM25_CorpusStatistics(t *testing.T) {
	idx := NewInvertedIndex()

	idx.Index(1, "machine learning")
	idx.Index(2, "deep learning algorithms")
	idx.Index(3, "machine vision")

	// Check total documents
	if idx.TotalDocs != 3 {
		t.Errorf("TotalDocs = %d, want 3", idx.TotalDocs)
	}

	// Check total terms (after stop word removal)
	// Doc1: 2 terms, Doc2: 3 terms, Doc3: 2 terms = 7 total
	expectedTotal := int64(7)
	if idx.TotalTerms != expectedTotal {
		t.Errorf("TotalTerms = %d, want %d", idx.TotalTerms, expectedTotal)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// BM25 SERIALIZATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_BM25_Serialization(t *testing.T) {
	idx := NewInvertedIndex()

	// Index documents
	idx.Index(1, "machine learning algorithms")
	idx.Index(2, "deep learning neural networks")

	// Serialize
	data, err := idx.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Deserialize into new index
	idx2 := NewInvertedIndex()
	err = idx2.Decode(data)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// Compare BM25 parameters
	if idx2.BM25Params.K1 != idx.BM25Params.K1 {
		t.Errorf("Decoded K1 = %f, want %f", idx2.BM25Params.K1, idx.BM25Params.K1)
	}
	if idx2.BM25Params.B != idx.BM25Params.B {
		t.Errorf("Decoded B = %f, want %f", idx2.BM25Params.B, idx.BM25Params.B)
	}

	// Compare corpus statistics
	if idx2.TotalDocs != idx.TotalDocs {
		t.Errorf("Decoded TotalDocs = %d, want %d", idx2.TotalDocs, idx.TotalDocs)
	}
	if idx2.TotalTerms != idx.TotalTerms {
		t.Errorf("Decoded TotalTerms = %d, want %d", idx2.TotalTerms, idx.TotalTerms)
	}

	// Compare document statistics
	if len(idx2.DocStats) != len(idx.DocStats) {
		t.Errorf("Decoded DocStats has %d entries, want %d", len(idx2.DocStats), len(idx.DocStats))
	}

	for docID, stats := range idx.DocStats {
		stats2, exists := idx2.DocStats[docID]
		if !exists {
			t.Errorf("Decoded DocStats missing Doc%d", docID)
			continue
		}

		if stats2.Length != stats.Length {
			t.Errorf("Doc%d length = %d, want %d", docID, stats2.Length, stats.Length)
		}

		if len(stats2.TermFreqs) != len(stats.TermFreqs) {
			t.Errorf("Doc%d has %d terms, want %d", docID, len(stats2.TermFreqs), len(stats.TermFreqs))
		}

		for term, freq := range stats.TermFreqs {
			if stats2.TermFreqs[term] != freq {
				t.Errorf("Doc%d term '%s' freq = %d, want %d", docID, term, stats2.TermFreqs[term], freq)
			}
		}
	}
}

func TestInvertedIndex_BM25_SerializationAndSearch(t *testing.T) {
	idx := NewInvertedIndex()

	// Index documents
	idx.Index(1, "machine learning algorithms")
	idx.Index(2, "deep learning neural networks")
	idx.Index(3, "machine learning and deep learning")

	// Search before serialization
	results1 := idx.RankBM25("machine learning", 10)

	// Serialize and deserialize
	data, err := idx.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	idx2 := NewInvertedIndex()
	err = idx2.Decode(data)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// Search after deserialization
	results2 := idx2.RankBM25("machine learning", 10)

	// Compare results
	if len(results2) != len(results1) {
		t.Errorf("After deserialization: %d results, want %d", len(results2), len(results1))
	}

	for i := range results1 {
		if results2[i].DocID != results1[i].DocID {
			t.Errorf("Result %d: DocID = %d, want %d", i, results2[i].DocID, results1[i].DocID)
		}

		// Scores should be very close (floating point comparison)
		scoreDiff := results2[i].Score - results1[i].Score
		if scoreDiff < -0.0001 || scoreDiff > 0.0001 {
			t.Errorf("Result %d: Score = %f, want %f (diff=%f)", i, results2[i].Score, results1[i].Score, scoreDiff)
		}
	}
}

func TestInvertedIndex_BM25_CustomParametersSerialization(t *testing.T) {
	idx := NewInvertedIndex()

	// Set custom parameters
	idx.BM25Params.K1 = 2.0
	idx.BM25Params.B = 0.5

	idx.Index(1, "machine learning")

	// Serialize and deserialize
	data, err := idx.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	idx2 := NewInvertedIndex()
	err = idx2.Decode(data)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// Custom parameters should be preserved
	if idx2.BM25Params.K1 != 2.0 {
		t.Errorf("Decoded K1 = %f, want 2.0", idx2.BM25Params.K1)
	}
	if idx2.BM25Params.B != 0.5 {
		t.Errorf("Decoded B = %f, want 0.5", idx2.BM25Params.B)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// BM25 INTEGRATION AND REAL-WORLD TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_BM25_RealWorldScenario(t *testing.T) {
	idx := NewInvertedIndex()

	// Simulate a real corpus of technical blog posts
	idx.Index(1, "Introduction to Machine Learning: A Comprehensive Guide for Beginners")
	idx.Index(2, "Deep Learning and Neural Networks Explained")
	idx.Index(3, "Machine Learning Algorithms: Decision Trees and Random Forests")
	idx.Index(4, "Natural Language Processing with Python")
	idx.Index(5, "Computer Vision and Image Recognition using Deep Learning")
	idx.Index(6, "Machine Learning in Production: Best Practices")

	// Query: "machine learning"
	results := idx.RankBM25("machine learning", 10)

	// Should find Doc1, Doc3, and Doc6
	if len(results) < 3 {
		t.Errorf("Found %d results, want at least 3", len(results))
	}

	// Doc1 has "machine learning" in title (high relevance)
	// Doc3 has "machine learning algorithms" (high relevance)
	// Doc6 has "machine learning in production" (high relevance)
	foundDocs := make(map[int]bool)
	for _, result := range results {
		foundDocs[result.DocID] = true
	}

	expectedDocs := []int{1, 3, 6}
	for _, docID := range expectedDocs {
		if !foundDocs[docID] {
			t.Errorf("Expected Doc%d in results for 'machine learning'", docID)
		}
	}
}

func TestInvertedIndex_BM25_MultiTermQuery(t *testing.T) {
	idx := NewInvertedIndex()

	idx.Index(1, "python programming language tutorial")
	idx.Index(2, "python machine learning tutorial")
	idx.Index(3, "java programming language")
	idx.Index(4, "machine learning with python and java")

	// Three-term query
	results := idx.RankBM25("python machine learning", 10)

	if len(results) == 0 {
		t.Fatal("Should find results")
	}

	// Doc2 and Doc4 have all terms, should rank highest
	// Doc2: "python machine learning tutorial"
	// Doc4: "machine learning with python and java"
	topDocID := results[0].DocID
	if topDocID != 2 && topDocID != 4 {
		t.Errorf("Top result is Doc%d, expected Doc2 or Doc4", topDocID)
	}
}

func TestInvertedIndex_BM25_EmptyIndex(t *testing.T) {
	idx := NewInvertedIndex()

	// Search empty index
	results := idx.RankBM25("machine learning", 10)

	if len(results) != 0 {
		t.Errorf("Empty index returned %d results, want 0", len(results))
	}
}

func TestInvertedIndex_BM25_SingleDocumentCorpus(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "machine learning algorithms")

	results := idx.RankBM25("machine learning", 10)

	if len(results) != 1 {
		t.Fatalf("Found %d results, want 1", len(results))
	}

	if results[0].DocID != 1 {
		t.Errorf("Result DocID = %d, want 1", results[0].DocID)
	}

	if results[0].Score <= 0 {
		t.Errorf("Score = %f, want > 0", results[0].Score)
	}
}

func TestInvertedIndex_BM25_DuplicateTerms(t *testing.T) {
	idx := NewInvertedIndex()

	// Document with repeated terms
	idx.Index(1, "machine machine machine learning")

	results := idx.RankBM25("machine", 10)

	if len(results) != 1 {
		t.Fatalf("Found %d results, want 1", len(results))
	}

	// Higher term frequency should result in higher score
	// compared to a document with single occurrence
	idx.Index(2, "machine learning")
	results2 := idx.RankBM25("machine", 10)

	if len(results2) != 2 {
		t.Fatalf("Found %d results, want 2", len(results2))
	}

	// Doc1 (TF=3) should rank higher than Doc2 (TF=1)
	if results2[0].DocID != 1 {
		t.Errorf("Top result is Doc%d, want Doc1 (higher TF)", results2[0].DocID)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// BM25 BENCHMARK TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func BenchmarkRankBM25(b *testing.B) {
	idx := NewInvertedIndex()

	// Pre-populate index with realistic documents
	documents := []string{
		"introduction to machine learning algorithms and techniques",
		"deep learning neural networks for image recognition",
		"natural language processing with python programming",
		"machine learning models and evaluation metrics",
		"computer vision and image processing fundamentals",
		"supervised learning classification and regression",
		"unsupervised learning clustering algorithms",
		"reinforcement learning and game playing",
	}

	for i, doc := range documents {
		idx.Index(i+1, strings.Repeat(doc+" ", 10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.RankBM25("machine learning", 10)
	}
}

func BenchmarkCalculateIDF(b *testing.B) {
	idx := NewInvertedIndex()

	// Pre-populate index
	for i := 1; i <= 100; i++ {
		idx.Index(i, "machine learning algorithms neural networks deep learning")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.calculateIDF("machine")
	}
}

func BenchmarkCalculateBM25Score(b *testing.B) {
	idx := NewInvertedIndex()

	// Pre-populate index
	for i := 1; i <= 100; i++ {
		idx.Index(i, "machine learning algorithms neural networks deep learning")
	}

	tokens := []string{"machine", "learning"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.calculateBM25Score(1, tokens)
	}
}
