package blaze

import (
	"math"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════════════
// POSITION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestPosition_GetDocumentID(t *testing.T) {
	pos := Position{DocumentID: 42, Offset: 10}
	if got := pos.GetDocumentID(); got != 42 {
		t.Errorf("GetDocumentID() = %d, want 42", got)
	}
}

func TestPosition_GetOffset(t *testing.T) {
	pos := Position{DocumentID: 42, Offset: 10}
	if got := pos.GetOffset(); got != 10 {
		t.Errorf("GetOffset() = %d, want 10", got)
	}
}

func TestPosition_IsBeginning(t *testing.T) {
	tests := []struct {
		name string
		pos  Position
		want bool
	}{
		{"BOF position", Position{DocumentID: BOF, Offset: BOF}, true},
		{"Regular position", Position{DocumentID: 1, Offset: 0}, false},
		{"EOF position", Position{DocumentID: EOF, Offset: EOF}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.IsBeginning(); got != tt.want {
				t.Errorf("IsBeginning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPosition_IsEnd(t *testing.T) {
	tests := []struct {
		name string
		pos  Position
		want bool
	}{
		{"EOF position", Position{DocumentID: EOF, Offset: EOF}, true},
		{"Regular position", Position{DocumentID: 1, Offset: 0}, false},
		{"BOF position", Position{DocumentID: BOF, Offset: BOF}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.IsEnd(); got != tt.want {
				t.Errorf("IsEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPosition_IsBefore(t *testing.T) {
	tests := []struct {
		name  string
		pos   Position
		other Position
		want  bool
	}{
		{
			"Same doc, earlier offset",
			Position{DocumentID: 1, Offset: 5},
			Position{DocumentID: 1, Offset: 10},
			true,
		},
		{
			"Same doc, later offset",
			Position{DocumentID: 1, Offset: 10},
			Position{DocumentID: 1, Offset: 5},
			false,
		},
		{
			"Earlier doc",
			Position{DocumentID: 1, Offset: 100},
			Position{DocumentID: 2, Offset: 0},
			true,
		},
		{
			"Later doc",
			Position{DocumentID: 2, Offset: 0},
			Position{DocumentID: 1, Offset: 100},
			false,
		},
		{
			"BOF before regular",
			Position{DocumentID: BOF, Offset: BOF},
			Position{DocumentID: 1, Offset: 0},
			true,
		},
		{
			"Regular before EOF",
			Position{DocumentID: 1, Offset: 0},
			Position{DocumentID: EOF, Offset: EOF},
			true,
		},
		{
			"Same position",
			Position{DocumentID: 1, Offset: 5},
			Position{DocumentID: 1, Offset: 5},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.IsBefore(tt.other); got != tt.want {
				t.Errorf("IsBefore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPosition_IsAfter(t *testing.T) {
	tests := []struct {
		name  string
		pos   Position
		other Position
		want  bool
	}{
		{
			"Same doc, later offset",
			Position{DocumentID: 1, Offset: 10},
			Position{DocumentID: 1, Offset: 5},
			true,
		},
		{
			"Same doc, earlier offset",
			Position{DocumentID: 1, Offset: 5},
			Position{DocumentID: 1, Offset: 10},
			false,
		},
		{
			"Later doc",
			Position{DocumentID: 2, Offset: 0},
			Position{DocumentID: 1, Offset: 100},
			true,
		},
		{
			"Earlier doc",
			Position{DocumentID: 1, Offset: 100},
			Position{DocumentID: 2, Offset: 0},
			false,
		},
		{
			"EOF after regular",
			Position{DocumentID: EOF, Offset: EOF},
			Position{DocumentID: 1, Offset: 0},
			true,
		},
		{
			"Regular after BOF",
			Position{DocumentID: 1, Offset: 0},
			Position{DocumentID: BOF, Offset: BOF},
			true,
		},
		{
			"Same position",
			Position{DocumentID: 1, Offset: 5},
			Position{DocumentID: 1, Offset: 5},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.IsAfter(tt.other); got != tt.want {
				t.Errorf("IsAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPosition_Equals(t *testing.T) {
	tests := []struct {
		name  string
		pos   Position
		other Position
		want  bool
	}{
		{
			"Same position",
			Position{DocumentID: 1, Offset: 5},
			Position{DocumentID: 1, Offset: 5},
			true,
		},
		{
			"Different offset",
			Position{DocumentID: 1, Offset: 5},
			Position{DocumentID: 1, Offset: 10},
			false,
		},
		{
			"Different document",
			Position{DocumentID: 1, Offset: 5},
			Position{DocumentID: 2, Offset: 5},
			false,
		},
		{
			"Both BOF",
			Position{DocumentID: BOF, Offset: BOF},
			Position{DocumentID: BOF, Offset: BOF},
			true,
		},
		{
			"Both EOF",
			Position{DocumentID: EOF, Offset: EOF},
			Position{DocumentID: EOF, Offset: EOF},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.Equals(tt.other); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// SKIP LIST BASIC TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestNewSkipList(t *testing.T) {
	sl := NewSkipList()

	if sl.Head == nil {
		t.Error("NewSkipList() created nil Head")
	}

	if sl.Height != 1 {
		t.Errorf("NewSkipList() Height = %d, want 1", sl.Height)
	}
}

func TestSkipList_Insert_Single(t *testing.T) {
	sl := NewSkipList()
	pos := Position{DocumentID: 1, Offset: 5}

	sl.Insert(pos)

	// Verify the element was inserted
	found, err := sl.Find(pos)
	if err != nil {
		t.Errorf("Find() error = %v, want nil", err)
	}

	if !found.Equals(pos) {
		t.Errorf("Find() = %v, want %v", found, pos)
	}
}

func TestSkipList_Insert_Multiple(t *testing.T) {
	sl := NewSkipList()

	positions := []Position{
		{DocumentID: 1, Offset: 5},
		{DocumentID: 1, Offset: 10},
		{DocumentID: 2, Offset: 0},
		{DocumentID: 2, Offset: 15},
		{DocumentID: 3, Offset: 7},
	}

	// Insert all positions
	for _, pos := range positions {
		sl.Insert(pos)
	}

	// Verify all can be found
	for _, pos := range positions {
		found, err := sl.Find(pos)
		if err != nil {
			t.Errorf("Find(%v) error = %v, want nil", pos, err)
		}
		if !found.Equals(pos) {
			t.Errorf("Find(%v) = %v, want %v", pos, found, pos)
		}
	}
}

func TestSkipList_Insert_Duplicate(t *testing.T) {
	sl := NewSkipList()
	pos := Position{DocumentID: 1, Offset: 5}

	// Insert twice
	sl.Insert(pos)
	sl.Insert(pos)

	// Should only exist once
	found, err := sl.Find(pos)
	if err != nil {
		t.Errorf("Find() error = %v, want nil", err)
	}
	if !found.Equals(pos) {
		t.Errorf("Find() = %v, want %v", found, pos)
	}

	// Count elements using iterator
	count := 0
	iter := sl.Iterator()
	// First element is at current position
	if iter.current != nil {
		count++
	}
	// Rest of elements via HasNext/Next
	for iter.HasNext() {
		iter.Next()
		count++
	}

	if count != 1 {
		t.Errorf("Skip list has %d elements, want 1", count)
	}
}

func TestSkipList_Insert_OutOfOrder(t *testing.T) {
	sl := NewSkipList()

	// Insert in reverse order
	positions := []Position{
		{DocumentID: 5, Offset: 10},
		{DocumentID: 3, Offset: 7},
		{DocumentID: 4, Offset: 2},
		{DocumentID: 1, Offset: 0},
		{DocumentID: 2, Offset: 5},
	}

	for _, pos := range positions {
		sl.Insert(pos)
	}

	// Verify they're stored in sorted order
	expected := []Position{
		{DocumentID: 1, Offset: 0},
		{DocumentID: 2, Offset: 5},
		{DocumentID: 3, Offset: 7},
		{DocumentID: 4, Offset: 2},
		{DocumentID: 5, Offset: 10},
	}

	// Get all positions using iterator
	var result []Position
	iter := sl.Iterator()
	// Get first element
	if iter.current != nil {
		result = append(result, iter.current.Key)
	}
	// Get remaining elements
	for iter.HasNext() {
		result = append(result, iter.Next())
	}

	if len(result) != len(expected) {
		t.Fatalf("Got %d positions, want %d", len(result), len(expected))
	}

	for idx, pos := range result {
		if !pos.Equals(expected[idx]) {
			t.Errorf("Position at index %d = %v, want %v", idx, pos, expected[idx])
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// SEARCH AND FIND TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestSkipList_Find_NotFound(t *testing.T) {
	sl := NewSkipList()
	sl.Insert(Position{DocumentID: 1, Offset: 5})
	sl.Insert(Position{DocumentID: 2, Offset: 10})

	pos := Position{DocumentID: 1, Offset: 7}
	_, err := sl.Find(pos)

	if err != ErrKeyNotFound {
		t.Errorf("Find() error = %v, want %v", err, ErrKeyNotFound)
	}
}

func TestSkipList_Find_EmptyList(t *testing.T) {
	sl := NewSkipList()

	pos := Position{DocumentID: 1, Offset: 0}
	_, err := sl.Find(pos)

	if err != ErrKeyNotFound {
		t.Errorf("Find() error = %v, want %v", err, ErrKeyNotFound)
	}
}

func TestSkipList_FindLessThan(t *testing.T) {
	sl := NewSkipList()

	// Insert: 5, 10, 15, 20
	sl.Insert(Position{DocumentID: 1, Offset: 5})
	sl.Insert(Position{DocumentID: 1, Offset: 10})
	sl.Insert(Position{DocumentID: 1, Offset: 15})
	sl.Insert(Position{DocumentID: 1, Offset: 20})

	tests := []struct {
		name    string
		key     Position
		want    Position
		wantErr error
	}{
		{
			"Find less than 17",
			Position{DocumentID: 1, Offset: 17},
			Position{DocumentID: 1, Offset: 15},
			nil,
		},
		{
			"Find less than 15",
			Position{DocumentID: 1, Offset: 15},
			Position{DocumentID: 1, Offset: 10},
			nil,
		},
		{
			"Find less than 5 (first element)",
			Position{DocumentID: 1, Offset: 5},
			BOFDocument,
			ErrNoElementFound,
		},
		{
			"Find less than 0 (before first)",
			Position{DocumentID: 1, Offset: 0},
			BOFDocument,
			ErrNoElementFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sl.FindLessThan(tt.key)

			if err != tt.wantErr {
				t.Errorf("FindLessThan() error = %v, want %v", err, tt.wantErr)
			}

			if !got.Equals(tt.want) {
				t.Errorf("FindLessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkipList_FindGreaterThan(t *testing.T) {
	sl := NewSkipList()

	// Insert: 5, 10, 15, 20
	sl.Insert(Position{DocumentID: 1, Offset: 5})
	sl.Insert(Position{DocumentID: 1, Offset: 10})
	sl.Insert(Position{DocumentID: 1, Offset: 15})
	sl.Insert(Position{DocumentID: 1, Offset: 20})

	tests := []struct {
		name    string
		key     Position
		want    Position
		wantErr error
	}{
		{
			"Find greater than 10 (exists)",
			Position{DocumentID: 1, Offset: 10},
			Position{DocumentID: 1, Offset: 15},
			nil,
		},
		{
			"Find greater than 12 (doesn't exist)",
			Position{DocumentID: 1, Offset: 12},
			Position{DocumentID: 1, Offset: 15},
			nil,
		},
		{
			"Find greater than 20 (last element)",
			Position{DocumentID: 1, Offset: 20},
			EOFDocument,
			ErrNoElementFound,
		},
		{
			"Find greater than 25 (after last)",
			Position{DocumentID: 1, Offset: 25},
			EOFDocument,
			ErrNoElementFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sl.FindGreaterThan(tt.key)

			if err != tt.wantErr {
				t.Errorf("FindGreaterThan() error = %v, want %v", err, tt.wantErr)
			}

			if !got.Equals(tt.want) {
				t.Errorf("FindGreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// DELETE TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestSkipList_Delete_Single(t *testing.T) {
	sl := NewSkipList()
	pos := Position{DocumentID: 1, Offset: 5}

	sl.Insert(pos)

	// Delete the element
	deleted := sl.Delete(pos)
	if !deleted {
		t.Error("Delete() = false, want true")
	}

	// Verify it's gone
	_, err := sl.Find(pos)
	if err != ErrKeyNotFound {
		t.Errorf("Find() after delete error = %v, want %v", err, ErrKeyNotFound)
	}
}

func TestSkipList_Delete_Middle(t *testing.T) {
	sl := NewSkipList()

	// Insert: 5, 10, 15, 20
	sl.Insert(Position{DocumentID: 1, Offset: 5})
	sl.Insert(Position{DocumentID: 1, Offset: 10})
	sl.Insert(Position{DocumentID: 1, Offset: 15})
	sl.Insert(Position{DocumentID: 1, Offset: 20})

	// Delete middle element
	deleted := sl.Delete(Position{DocumentID: 1, Offset: 10})
	if !deleted {
		t.Error("Delete() = false, want true")
	}

	// Verify it's gone
	_, err := sl.Find(Position{DocumentID: 1, Offset: 10})
	if err != ErrKeyNotFound {
		t.Errorf("Find() after delete error = %v, want %v", err, ErrKeyNotFound)
	}

	// Verify others still exist
	remaining := []Position{
		{DocumentID: 1, Offset: 5},
		{DocumentID: 1, Offset: 15},
		{DocumentID: 1, Offset: 20},
	}

	for _, pos := range remaining {
		_, err := sl.Find(pos)
		if err != nil {
			t.Errorf("Find(%v) error = %v, want nil", pos, err)
		}
	}
}

func TestSkipList_Delete_NotFound(t *testing.T) {
	sl := NewSkipList()
	sl.Insert(Position{DocumentID: 1, Offset: 5})

	deleted := sl.Delete(Position{DocumentID: 1, Offset: 10})
	if deleted {
		t.Error("Delete() = true, want false (element not found)")
	}
}

func TestSkipList_Delete_EmptyList(t *testing.T) {
	sl := NewSkipList()

	deleted := sl.Delete(Position{DocumentID: 1, Offset: 0})
	if deleted {
		t.Error("Delete() = true, want false (empty list)")
	}
}

func TestSkipList_Delete_All(t *testing.T) {
	sl := NewSkipList()

	positions := []Position{
		{DocumentID: 1, Offset: 5},
		{DocumentID: 1, Offset: 10},
		{DocumentID: 2, Offset: 15},
	}

	// Insert all
	for _, pos := range positions {
		sl.Insert(pos)
	}

	// Delete all
	for _, pos := range positions {
		deleted := sl.Delete(pos)
		if !deleted {
			t.Errorf("Delete(%v) = false, want true", pos)
		}
	}

	// Verify list is empty
	iter := sl.Iterator()
	if iter.HasNext() {
		t.Error("List should be empty after deleting all elements")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ITERATOR TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestSkipList_Iterator_Empty(t *testing.T) {
	sl := NewSkipList()
	iter := sl.Iterator()

	if iter.HasNext() {
		t.Error("HasNext() = true for empty list, want false")
	}

	pos := iter.Next()
	if !pos.Equals(EOFDocument) {
		t.Errorf("Next() on empty list = %v, want %v", pos, EOFDocument)
	}
}

func TestSkipList_Iterator_Single(t *testing.T) {
	sl := NewSkipList()
	expected := Position{DocumentID: 1, Offset: 5}
	sl.Insert(expected)

	iter := sl.Iterator()

	// First element is at current position
	if iter.current == nil {
		t.Fatal("Iterator current is nil, expected first element")
	}

	pos := iter.current.Key
	if !pos.Equals(expected) {
		t.Errorf("First element = %v, want %v", pos, expected)
	}

	// Should have no next elements
	if iter.HasNext() {
		t.Error("HasNext() = true for single element list, want false")
	}
}

func TestSkipList_Iterator_Multiple(t *testing.T) {
	sl := NewSkipList()

	expected := []Position{
		{DocumentID: 1, Offset: 5},
		{DocumentID: 1, Offset: 10},
		{DocumentID: 2, Offset: 0},
		{DocumentID: 2, Offset: 15},
		{DocumentID: 3, Offset: 7},
	}

	// Insert all
	for _, pos := range expected {
		sl.Insert(pos)
	}

	// Get all positions using iterator
	var result []Position
	iter := sl.Iterator()
	// Get first element
	if iter.current != nil {
		result = append(result, iter.current.Key)
	}
	// Get remaining elements
	for iter.HasNext() {
		result = append(result, iter.Next())
	}

	if len(result) != len(expected) {
		t.Errorf("Iterator returned %d elements, want %d", len(result), len(expected))
	}

	for idx, pos := range result {
		if idx >= len(expected) {
			t.Fatalf("Iterator returned more elements than expected")
		}

		if !pos.Equals(expected[idx]) {
			t.Errorf("Position at index %d = %v, want %v", idx, pos, expected[idx])
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// LAST OPERATION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestSkipList_Last_Empty(t *testing.T) {
	sl := NewSkipList()
	last := sl.Last()

	// In an empty list, Last() returns the head's key (which is zero value)
	if last.DocumentID != 0 || last.Offset != 0 {
		t.Errorf("Last() on empty list = %v, want zero position", last)
	}
}

func TestSkipList_Last_Single(t *testing.T) {
	sl := NewSkipList()
	expected := Position{DocumentID: 1, Offset: 5}
	sl.Insert(expected)

	last := sl.Last()
	if !last.Equals(expected) {
		t.Errorf("Last() = %v, want %v", last, expected)
	}
}

func TestSkipList_Last_Multiple(t *testing.T) {
	sl := NewSkipList()

	sl.Insert(Position{DocumentID: 1, Offset: 5})
	sl.Insert(Position{DocumentID: 2, Offset: 10})
	sl.Insert(Position{DocumentID: 3, Offset: 15})

	expected := Position{DocumentID: 3, Offset: 15}
	last := sl.Last()

	if !last.Equals(expected) {
		t.Errorf("Last() = %v, want %v", last, expected)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// EDGE CASE AND STRESS TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func TestSkipList_SameDocument_DifferentOffsets(t *testing.T) {
	sl := NewSkipList()

	// Insert multiple positions in the same document
	for offset := 0; offset < 10; offset++ {
		sl.Insert(Position{DocumentID: 1, Offset: offset})
	}

	// Verify all are present and in order
	var result []Position
	iter := sl.Iterator()
	// Get first element
	if iter.current != nil {
		result = append(result, iter.current.Key)
	}
	// Get remaining elements
	for iter.HasNext() {
		result = append(result, iter.Next())
	}

	if len(result) != 10 {
		t.Errorf("Found %d positions, want 10", len(result))
	}

	for offset, pos := range result {
		expected := Position{DocumentID: 1, Offset: offset}

		if !pos.Equals(expected) {
			t.Errorf("Position at offset %d = %v, want %v", offset, pos, expected)
		}
	}
}

func TestSkipList_MultipleDocs_MultipleOffsets(t *testing.T) {
	sl := NewSkipList()

	// Insert grid: 3 documents x 5 offsets each
	for doc := 1; doc <= 3; doc++ {
		for offset := 0; offset < 5; offset++ {
			sl.Insert(Position{DocumentID: doc, Offset: offset})
		}
	}

	// Get all positions using iterator
	var result []Position
	iter := sl.Iterator()
	// Get first element
	if iter.current != nil {
		result = append(result, iter.current.Key)
	}
	// Get remaining elements
	for iter.HasNext() {
		result = append(result, iter.Next())
	}

	if len(result) != 15 {
		t.Errorf("Found %d positions, want 15", len(result))
	}

	// Verify ordering (should be doc-major order)
	idx := 0
	for doc := 1; doc <= 3; doc++ {
		for offset := 0; offset < 5; offset++ {
			if idx >= len(result) {
				t.Fatal("Not enough positions in result")
			}

			expected := Position{DocumentID: doc, Offset: offset}

			if !result[idx].Equals(expected) {
				t.Errorf("Position at index %d = %v, want %v", idx, result[idx], expected)
			}

			idx++
		}
	}
}

func TestSkipList_LargeDataset(t *testing.T) {
	sl := NewSkipList()

	// Insert 1000 positions
	n := 1000
	for i := 0; i < n; i++ {
		sl.Insert(Position{DocumentID: i / 10, Offset: i % 10})
	}

	// Verify count
	count := 0
	iter := sl.Iterator()
	// Count first element
	if iter.current != nil {
		count++
	}
	// Count remaining elements
	for iter.HasNext() {
		iter.Next()
		count++
	}

	if count != n {
		t.Errorf("Found %d positions, want %d", count, n)
	}

	// Spot check some positions
	testPositions := []Position{
		{DocumentID: 0, Offset: 0},
		{DocumentID: 50, Offset: 5},
		{DocumentID: 99, Offset: 9},
	}

	for _, pos := range testPositions {
		found, err := sl.Find(pos)
		if err != nil {
			t.Errorf("Find(%v) error = %v, want nil", pos, err)
		}
		if !found.Equals(pos) {
			t.Errorf("Find(%v) = %v, want %v", pos, found, pos)
		}
	}
}

func TestSkipList_InfinityValues(t *testing.T) {
	// Test that sentinel values work correctly
	if BOF >= 0 {
		t.Error("BOF should be negative (math.MinInt)")
	}

	if EOF <= 0 {
		t.Error("EOF should be positive (math.MaxInt)")
	}

	if BOF != math.MinInt {
		t.Errorf("BOF should be math.MinInt, got %d", BOF)
	}

	if EOF != math.MaxInt {
		t.Errorf("EOF should be math.MaxInt, got %d", EOF)
	}

	// BOF should be less than any regular position
	regularPos := Position{DocumentID: 0, Offset: 0}
	if !BOFDocument.IsBefore(regularPos) {
		t.Error("BOF should be before any regular position")
	}

	// EOF should be greater than any regular position
	if !regularPos.IsBefore(EOFDocument) {
		t.Error("Any regular position should be before EOF")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// BENCHMARK TESTS
// ═══════════════════════════════════════════════════════════════════════════════

func BenchmarkSkipList_Insert(b *testing.B) {
	sl := NewSkipList()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.Insert(Position{DocumentID: i / 1000, Offset: i % 1000})
	}
}

func BenchmarkSkipList_Find(b *testing.B) {
	sl := NewSkipList()

	// Pre-populate with 10000 elements
	for i := 0; i < 10000; i++ {
		sl.Insert(Position{DocumentID: i / 100, Offset: i % 100})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.Find(Position{DocumentID: i / 100 % 100, Offset: i % 100})
	}
}

func BenchmarkSkipList_Delete(b *testing.B) {
	// Re-populate for each iteration
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		sl := NewSkipList()
		for j := 0; j < 1000; j++ {
			sl.Insert(Position{DocumentID: j / 10, Offset: j % 10})
		}
		b.StartTimer()

		sl.Delete(Position{DocumentID: i / 10 % 100, Offset: i % 10})
	}
}

func BenchmarkSkipList_Iterator(b *testing.B) {
	sl := NewSkipList()

	// Pre-populate with 1000 elements
	for i := 0; i < 1000; i++ {
		sl.Insert(Position{DocumentID: i / 10, Offset: i % 10})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := sl.Iterator()
		// Process first element
		if iter.current != nil {
			_ = iter.current.Key
		}
		// Process remaining elements
		for iter.HasNext() {
			iter.Next()
		}
	}
}

func BenchmarkSkipList_FindLessThan(b *testing.B) {
	sl := NewSkipList()

	// Pre-populate with 10000 elements
	for i := 0; i < 10000; i++ {
		sl.Insert(Position{DocumentID: i / 100, Offset: i % 100})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.FindLessThan(Position{DocumentID: i / 100 % 100, Offset: i % 100})
	}
}

func BenchmarkSkipList_FindGreaterThan(b *testing.B) {
	sl := NewSkipList()

	// Pre-populate with 10000 elements
	for i := 0; i < 10000; i++ {
		sl.Insert(Position{DocumentID: i / 100, Offset: i % 100})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl.FindGreaterThan(Position{DocumentID: i / 100 % 100, Offset: i % 100})
	}
}
