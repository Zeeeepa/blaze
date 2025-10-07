package blaze

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════════════
// INVERTED INDEX CREATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestNewInvertedIndex(t *testing.T) {
	idx := NewInvertedIndex()

	if idx == nil {
		t.Fatal("NewInvertedIndex() returned nil")
	}

	if idx.PostingsList == nil {
		t.Error("PostingsList is nil")
	}

	if len(idx.PostingsList) != 0 {
		t.Errorf("New index has %d entries, want 0", len(idx.PostingsList))
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// INDEXING TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_Index_SingleDocument(t *testing.T) {
	idx := NewInvertedIndex()

	// Index a simple document
	idx.Index(1, "quick brown fox")

	// Verify tokens were indexed
	tokens := []string{"quick", "brown", "fox"}
	for _, token := range tokens {
		if _, exists := idx.PostingsList[token]; !exists {
			t.Errorf("Token %q was not indexed", token)
		}
	}
}

func TestInvertedIndex_Index_MultipleDocuments(t *testing.T) {
	idx := NewInvertedIndex()

	// Index multiple documents
	idx.Index(1, "quick brown fox")
	idx.Index(2, "sleepy dog")
	idx.Index(3, "quick brown cats")

	// Check that all unique tokens are indexed (after stemming)
	expectedTokens := map[string]bool{
		"quick":  true,
		"brown":  true,
		"fox":    true,
		"sleepi": true, // stemmed from "sleepy"
		"dog":    true,
		"cat":    true, // stemmed from "cats"
	}

	for token := range expectedTokens {
		if _, exists := idx.PostingsList[token]; !exists {
			t.Errorf("Token %q was not indexed", token)
		}
	}
}

func TestInvertedIndex_Index_DuplicateWords(t *testing.T) {
	idx := NewInvertedIndex()

	// Index document with duplicate words
	idx.Index(1, "quick quick brown")

	// Verify "quick" has multiple positions
	skipList, exists := idx.PostingsList["quick"]
	if !exists {
		t.Fatal("Token 'quick' was not indexed")
	}

	// Count occurrences
	count := 0
	iter := skipList.Iterator()
	if iter.current != nil {
		count++
	}
	for iter.HasNext() {
		iter.Next()
		count++
	}

	if count != 2 {
		t.Errorf("Token 'quick' has %d occurrences, want 2", count)
	}
}

func TestInvertedIndex_Index_EmptyDocument(t *testing.T) {
	idx := NewInvertedIndex()

	// Index empty document
	idx.Index(1, "")

	// Should have no tokens
	if len(idx.PostingsList) != 0 {
		t.Errorf("Empty document created %d tokens, want 0", len(idx.PostingsList))
	}
}

func TestInvertedIndex_Index_StopWords(t *testing.T) {
	idx := NewInvertedIndex()

	// Index document with stop words
	idx.Index(1, "the quick brown fox")

	// "the" should be removed by analyzer
	if _, exists := idx.PostingsList["the"]; exists {
		t.Error("Stop word 'the' should not be indexed")
	}

	// Other words should exist
	if _, exists := idx.PostingsList["quick"]; !exists {
		t.Error("Token 'quick' should be indexed")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// FIRST OPERATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_First_SingleOccurrence(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	pos, err := idx.First("quick")
	if err != nil {
		t.Fatalf("First() error = %v, want nil", err)
	}

	if pos.GetDocumentID() != 1 {
		t.Errorf("First() document = %d, want 1", pos.GetDocumentID())
	}

	if pos.GetOffset() != 0 {
		t.Errorf("First() offset = %d, want 0", pos.GetOffset())
	}
}

func TestInvertedIndex_First_MultipleOccurrences(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "brown fox")
	idx.Index(2, "quick brown")
	idx.Index(3, "brown dog")

	pos, err := idx.First("brown")
	if err != nil {
		t.Fatalf("First() error = %v, want nil", err)
	}

	// Should return the first occurrence (Doc1, Pos0)
	if pos.GetDocumentID() != 1 || pos.GetOffset() != 0 {
		t.Errorf("First() = Doc%d:Pos%d, want Doc1:Pos0",
			pos.GetDocumentID(), pos.GetOffset())
	}
}

func TestInvertedIndex_First_NotFound(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	_, err := idx.First("elephant")
	if err != ErrNoPostingList {
		t.Errorf("First() error = %v, want %v", err, ErrNoPostingList)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// LAST OPERATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_Last_SingleOccurrence(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	pos, err := idx.Last("fox")
	if err != nil {
		t.Fatalf("Last() error = %v, want nil", err)
	}

	if pos.GetDocumentID() != 1 || pos.GetOffset() != 2 {
		t.Errorf("Last() = Doc%d:Pos%d, want Doc1:Pos2",
			pos.GetDocumentID(), pos.GetOffset())
	}
}

func TestInvertedIndex_Last_MultipleOccurrences(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "brown fox")
	idx.Index(2, "quick brown")
	idx.Index(3, "brown dog")

	pos, err := idx.Last("brown")
	if err != nil {
		t.Fatalf("Last() error = %v, want nil", err)
	}

	// Should return the last occurrence (Doc3, Pos0)
	if pos.GetDocumentID() != 3 || pos.GetOffset() != 0 {
		t.Errorf("Last() = Doc%d:Pos%d, want Doc3:Pos0",
			pos.GetDocumentID(), pos.GetOffset())
	}
}

func TestInvertedIndex_Last_NotFound(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	_, err := idx.Last("elephant")
	if err != ErrNoPostingList {
		t.Errorf("Last() error = %v, want %v", err, ErrNoPostingList)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// NEXT OPERATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_Next_FromBeginning(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	// Next from BOF should return First
	pos, err := idx.Next("quick", BOFDocument)
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}

	if pos.GetDocumentID() != 1 || pos.GetOffset() != 0 {
		t.Errorf("Next() = Doc%d:Pos%d, want Doc1:Pos0",
			pos.GetDocumentID(), pos.GetOffset())
	}
}

func TestInvertedIndex_Next_MultipleOccurrences(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")
	idx.Index(2, "quick dog")
	idx.Index(3, "lazy quick")

	// Get first occurrence
	pos1, _ := idx.Next("quick", BOFDocument)
	if pos1.GetDocumentID() != 1 {
		t.Errorf("First occurrence in Doc%d, want Doc1", pos1.GetDocumentID())
	}

	// Get second occurrence
	pos2, _ := idx.Next("quick", pos1)
	if pos2.GetDocumentID() != 2 {
		t.Errorf("Second occurrence in Doc%d, want Doc2", pos2.GetDocumentID())
	}

	// Get third occurrence
	pos3, _ := idx.Next("quick", pos2)
	if pos3.GetDocumentID() != 3 {
		t.Errorf("Third occurrence in Doc%d, want Doc3", pos3.GetDocumentID())
	}

	// No more occurrences
	pos4, _ := idx.Next("quick", pos3)
	if !pos4.IsEnd() {
		t.Error("Next() should return EOF after last occurrence")
	}
}

func TestInvertedIndex_Next_FromEOF(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	pos, _ := idx.Next("quick", EOFDocument)
	if !pos.IsEnd() {
		t.Error("Next() from EOF should return EOF")
	}
}

func TestInvertedIndex_Next_NotFound(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	_, err := idx.Next("elephant", BOFDocument)
	if err != ErrNoPostingList {
		t.Errorf("Next() error = %v, want %v", err, ErrNoPostingList)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// PREVIOUS OPERATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_Previous_FromEnd(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	// Previous from EOF should return Last
	pos, err := idx.Previous("fox", EOFDocument)
	if err != nil {
		t.Fatalf("Previous() error = %v, want nil", err)
	}

	if pos.GetDocumentID() != 1 || pos.GetOffset() != 2 {
		t.Errorf("Previous() = Doc%d:Pos%d, want Doc1:Pos2",
			pos.GetDocumentID(), pos.GetOffset())
	}
}

func TestInvertedIndex_Previous_MultipleOccurrences(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")
	idx.Index(2, "quick dog")
	idx.Index(3, "lazy quick")

	// Get last occurrence
	pos3, _ := idx.Previous("quick", EOFDocument)
	if pos3.GetDocumentID() != 3 {
		t.Errorf("Last occurrence in Doc%d, want Doc3", pos3.GetDocumentID())
	}

	// Get second-to-last occurrence
	pos2, _ := idx.Previous("quick", pos3)
	if pos2.GetDocumentID() != 2 {
		t.Errorf("Second-to-last occurrence in Doc%d, want Doc2", pos2.GetDocumentID())
	}

	// Get first occurrence
	pos1, _ := idx.Previous("quick", pos2)
	if pos1.GetDocumentID() != 1 {
		t.Errorf("First occurrence in Doc%d, want Doc1", pos1.GetDocumentID())
	}

	// No more occurrences
	pos0, _ := idx.Previous("quick", pos1)
	if !pos0.IsBeginning() {
		t.Error("Previous() should return BOF before first occurrence")
	}
}

func TestInvertedIndex_Previous_FromBOF(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	pos, _ := idx.Previous("quick", BOFDocument)
	if !pos.IsBeginning() {
		t.Error("Previous() from BOF should return BOF")
	}
}

func TestInvertedIndex_Previous_NotFound(t *testing.T) {
	idx := NewInvertedIndex()
	idx.Index(1, "quick brown fox")

	_, err := idx.Previous("elephant", EOFDocument)
	if err != ErrNoPostingList {
		t.Errorf("Previous() error = %v, want %v", err, ErrNoPostingList)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// INTEGRATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_ComplexScenario(t *testing.T) {
	idx := NewInvertedIndex()

	// Index multiple documents with overlapping vocabulary
	idx.Index(1, "the quick brown fox jumps over the lazy dog")
	idx.Index(2, "the lazy brown dog sleeps")
	idx.Index(3, "quick brown foxes are clever")

	// Test 1: Verify "brown" appears in all three documents
	brownDocs := []int{}
	pos, _ := idx.First("brown")
	brownDocs = append(brownDocs, pos.GetDocumentID())

	for !pos.IsEnd() {
		pos, _ = idx.Next("brown", pos)
		if !pos.IsEnd() {
			brownDocs = append(brownDocs, pos.GetDocumentID())
		}
	}

	expectedDocs := []int{1, 2, 3}
	if len(brownDocs) != len(expectedDocs) {
		t.Errorf("Found 'brown' in %d documents, want %d", len(brownDocs), len(expectedDocs))
	}

	for i, docID := range brownDocs {
		if docID != expectedDocs[i] {
			t.Errorf("Document %d: got Doc%d, want Doc%d", i, docID, expectedDocs[i])
		}
	}

	// Test 2: Verify "quick" only appears in Doc1 and Doc3
	quickDocs := []int{}
	pos, _ = idx.First("quick")
	quickDocs = append(quickDocs, pos.GetDocumentID())

	pos, _ = idx.Next("quick", pos)
	if !pos.IsEnd() {
		quickDocs = append(quickDocs, pos.GetDocumentID())
	}

	expectedQuickDocs := []int{1, 3}
	if len(quickDocs) != len(expectedQuickDocs) {
		t.Errorf("Found 'quick' in %d documents, want %d", len(quickDocs), len(expectedQuickDocs))
	}
}

func TestInvertedIndex_PositionOrdering(t *testing.T) {
	idx := NewInvertedIndex()

	// Index document where same word appears multiple times
	idx.Index(1, "fox fox fox")

	// Get all positions
	var positions []int
	pos, _ := idx.First("fox")
	positions = append(positions, pos.GetOffset())

	for !pos.IsEnd() {
		pos, _ = idx.Next("fox", pos)
		if !pos.IsEnd() {
			positions = append(positions, pos.GetOffset())
		}
	}

	// Verify positions are in order: 0, 1, 2
	expected := []int{0, 1, 2}
	if len(positions) != len(expected) {
		t.Fatalf("Found %d positions, want %d", len(positions), len(expected))
	}

	for i, offset := range positions {
		if offset != expected[i] {
			t.Errorf("Position %d: offset = %d, want %d", i, offset, expected[i])
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// CONCURRENCY TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestInvertedIndex_ConcurrentIndexing(t *testing.T) {
	idx := NewInvertedIndex()

	// Index documents concurrently
	done := make(chan bool, 3)

	go func() {
		idx.Index(1, "quick brown fox")
		done <- true
	}()

	go func() {
		idx.Index(2, "sleepy dog")
		done <- true
	}()

	go func() {
		idx.Index(3, "quick brown cats")
		done <- true
	}()

	// Wait for all goroutines to complete
	<-done
	<-done
	<-done

	// Verify all documents were indexed (checking stemmed tokens)
	tokens := []string{"quick", "brown", "fox", "sleepi", "dog", "cat"}
	for _, token := range tokens {
		if _, exists := idx.PostingsList[token]; !exists {
			t.Errorf("Token %q was not indexed (concurrent indexing issue)", token)
		}
	}
}
