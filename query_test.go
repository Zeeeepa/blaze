package blaze

import (
	"testing"

	"github.com/RoaringBitmap/roaring"
)

// ═══════════════════════════════════════════════════════════════════════════════
// QUERY BUILDER TESTS
// ═══════════════════════════════════════════════════════════════════════════════

// setupTestIndex creates a test index with sample documents
func setupTestIndex() *InvertedIndex {
	idx := NewInvertedIndex()

	// Document 1: "machine learning is fun"
	idx.Index(1, "machine learning is fun")

	// Document 2: "deep learning and machine learning"
	idx.Index(2, "deep learning and machine learning")

	// Document 3: "python programming is great"
	idx.Index(3, "python programming is great")

	// Document 4: "machine learning with python"
	idx.Index(4, "machine learning with python")

	// Document 5: "cats and dogs are pets"
	idx.Index(5, "cats and dogs are pets")

	return idx
}

// TestQueryBuilder_SingleTerm tests querying for a single term
func TestQueryBuilder_SingleTerm(t *testing.T) {
	idx := setupTestIndex()

	// Query: Find documents with "machine"
	results := NewQueryBuilder(idx).
		Term("machine").
		Execute()

	// Should match docs 1, 2, 4
	expected := []int{1, 2, 4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_And tests AND operation
func TestQueryBuilder_And(t *testing.T) {
	idx := setupTestIndex()

	// Query: Find documents with "machine" AND "python"
	results := NewQueryBuilder(idx).
		Term("machine").
		And().
		Term("python").
		Execute()

	// Should match only doc 4
	expected := []int{4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_Or tests OR operation
func TestQueryBuilder_Or(t *testing.T) {
	idx := setupTestIndex()

	// Query: Find documents with "cats" OR "dogs"
	results := NewQueryBuilder(idx).
		Term("cats").
		Or().
		Term("dogs").
		Execute()

	// Should match doc 5 (which contains both)
	expected := []int{5}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_Not tests NOT operation
func TestQueryBuilder_Not(t *testing.T) {
	idx := setupTestIndex()

	// Query: Find documents with "learning" but NOT "deep"
	results := NewQueryBuilder(idx).
		Term("learning").
		And().Not().
		Term("deep").
		Execute()

	// Should match docs 1, 4 (not 2, which has "deep")
	expected := []int{1, 4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_ComplexQuery tests a complex boolean query
func TestQueryBuilder_ComplexQuery(t *testing.T) {
	idx := setupTestIndex()

	// Query: (machine OR python) AND learning
	results := NewQueryBuilder(idx).
		Group(func(q *QueryBuilder) {
			q.Term("machine").Or().Term("python")
		}).
		And().
		Term("learning").
		Execute()

	// Should match docs 1, 2, 4
	// Doc 1: has machine and learning
	// Doc 2: has machine and learning
	// Doc 3: has python but no learning
	// Doc 4: has machine, python, and learning
	expected := []int{1, 2, 4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_Phrase tests phrase query
func TestQueryBuilder_Phrase(t *testing.T) {
	idx := setupTestIndex()

	// Query: Find exact phrase "machine learning"
	results := NewQueryBuilder(idx).
		Phrase("machine learning").
		Execute()

	// Should match docs 1, 2, 4
	expected := []int{1, 2, 4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_PhraseWithBoolean tests combining phrase and boolean
func TestQueryBuilder_PhraseWithBoolean(t *testing.T) {
	idx := setupTestIndex()

	// Query: "machine learning" AND python
	results := NewQueryBuilder(idx).
		Phrase("machine learning").
		And().
		Term("python").
		Execute()

	// Should match only doc 4
	expected := []int{4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_ExecuteWithBM25 tests BM25 scoring
func TestQueryBuilder_ExecuteWithBM25(t *testing.T) {
	idx := setupTestIndex()

	// Query: machine AND learning (with BM25 scoring)
	results := NewQueryBuilder(idx).
		Term("machine").
		And().
		Term("learning").
		ExecuteWithBM25(10)

	// Should return docs with positive scores
	if len(results) == 0 {
		t.Error("Expected BM25 results, got none")
	}

	// All results should have positive scores
	for _, match := range results {
		if match.Score <= 0 {
			t.Errorf("Expected positive score, got %f", match.Score)
		}
	}

	// Results should be sorted by score (descending)
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("Results not sorted: score[%d]=%.2f > score[%d]=%.2f",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

// TestQueryBuilder_EmptyQuery tests empty query
func TestQueryBuilder_EmptyQuery(t *testing.T) {
	idx := setupTestIndex()

	// Empty query should return no results
	results := NewQueryBuilder(idx).Execute()

	if results.GetCardinality() != 0 {
		t.Errorf("Expected 0 results for empty query, got %d", results.GetCardinality())
	}
}

// TestQueryBuilder_NonExistentTerm tests querying for non-existent term
func TestQueryBuilder_NonExistentTerm(t *testing.T) {
	idx := setupTestIndex()

	// Query for a term that doesn't exist
	results := NewQueryBuilder(idx).
		Term("quantum").
		Execute()

	if results.GetCardinality() != 0 {
		t.Errorf("Expected 0 results for non-existent term, got %d", results.GetCardinality())
	}
}

// TestQueryBuilder_MultipleAnds tests chaining multiple AND operations
func TestQueryBuilder_MultipleAnds(t *testing.T) {
	idx := setupTestIndex()

	// Query: machine AND learning AND python
	results := NewQueryBuilder(idx).
		Term("machine").
		And().Term("learning").
		And().Term("python").
		Execute()

	// Should match only doc 4
	expected := []int{4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_MultipleOrs tests chaining multiple OR operations
func TestQueryBuilder_MultipleOrs(t *testing.T) {
	idx := setupTestIndex()

	// Query: cats OR dogs OR pets
	results := NewQueryBuilder(idx).
		Term("cats").
		Or().Term("dogs").
		Or().Term("pets").
		Execute()

	// Should match doc 5
	expected := []int{5}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_NestedGroups tests nested group operations
func TestQueryBuilder_NestedGroups(t *testing.T) {
	idx := setupTestIndex()

	// Query: ((machine OR deep) AND learning) AND NOT python
	results := NewQueryBuilder(idx).
		Group(func(q *QueryBuilder) {
			q.Group(func(qq *QueryBuilder) {
				qq.Term("machine").Or().Term("deep")
			}).And().Term("learning")
		}).
		And().Not().Term("python").
		Execute()

	// Should match docs 1, 2 (not 4 which has python)
	expected := []int{1, 2}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// CONVENIENCE FUNCTION TESTS
// ═══════════════════════════════════════════════════════════════════════════════

// TestAllOf tests AllOf convenience function
func TestAllOf(t *testing.T) {
	idx := setupTestIndex()

	// Find docs with machine, learning, and python
	results := AllOf(idx, "machine", "learning", "python")

	expected := []int{4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestAnyOf tests AnyOf convenience function
func TestAnyOf(t *testing.T) {
	idx := setupTestIndex()

	// Find docs with cats, dogs, or python
	results := AnyOf(idx, "cats", "dogs", "python")

	// Should match docs 3, 4, 5
	expected := []int{3, 4, 5}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestTermExcluding tests TermExcluding convenience function
func TestTermExcluding(t *testing.T) {
	idx := setupTestIndex()

	// Find docs with "learning" but not "deep"
	results := TermExcluding(idx, "learning", "deep")

	expected := []int{1, 4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestAllOf_EmptyTerms tests AllOf with no terms
func TestAllOf_EmptyTerms(t *testing.T) {
	idx := setupTestIndex()

	results := AllOf(idx)

	if results.GetCardinality() != 0 {
		t.Errorf("Expected 0 results for empty AllOf, got %d", results.GetCardinality())
	}
}

// TestAnyOf_EmptyTerms tests AnyOf with no terms
func TestAnyOf_EmptyTerms(t *testing.T) {
	idx := setupTestIndex()

	results := AnyOf(idx)

	if results.GetCardinality() != 0 {
		t.Errorf("Expected 0 results for empty AnyOf, got %d", results.GetCardinality())
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// REAL-WORLD QUERY PATTERNS
// ═══════════════════════════════════════════════════════════════════════════════

// TestQueryBuilder_SearchEnginePattern tests a typical search engine query
func TestQueryBuilder_SearchEnginePattern(t *testing.T) {
	idx := setupTestIndex()

	// Typical search: "machine learning" (phrase) OR just "python"
	results := NewQueryBuilder(idx).
		Phrase("machine learning").
		Or().
		Term("python").
		Execute()

	// Should match docs 1, 2, 3, 4
	expected := []int{1, 2, 3, 4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// TestQueryBuilder_FilteringPattern tests filtering unwanted content
func TestQueryBuilder_FilteringPattern(t *testing.T) {
	idx := setupTestIndex()

	// Find programming content but exclude python
	results := NewQueryBuilder(idx).
		Term("programming").
		And().Not().
		Term("python").
		Execute()

	// Should return no results (all programming docs have python)
	if results.GetCardinality() != 0 {
		t.Errorf("Expected 0 results, got %d", results.GetCardinality())
	}
}

// TestQueryBuilder_CategoryPattern tests category-based search
func TestQueryBuilder_CategoryPattern(t *testing.T) {
	idx := setupTestIndex()

	// Find AI/ML docs: (machine OR deep) AND learning
	results := NewQueryBuilder(idx).
		Group(func(q *QueryBuilder) {
			q.Term("machine").Or().Term("deep")
		}).
		And().
		Term("learning").
		Execute()

	expected := []int{1, 2, 4}
	actual := bitmapToSlice(results)

	if !slicesEqual(actual, expected) {
		t.Errorf("Expected docs %v, got %v", expected, actual)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// PERFORMANCE TESTS
// ═══════════════════════════════════════════════════════════════════════════════

// BenchmarkQueryBuilder_Simple benchmarks simple query
func BenchmarkQueryBuilder_Simple(b *testing.B) {
	idx := setupTestIndex()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewQueryBuilder(idx).
			Term("machine").
			And().
			Term("learning").
			Execute()
	}
}

// BenchmarkQueryBuilder_Complex benchmarks complex query
func BenchmarkQueryBuilder_Complex(b *testing.B) {
	idx := setupTestIndex()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewQueryBuilder(idx).
			Group(func(q *QueryBuilder) {
				q.Term("machine").Or().Term("deep")
			}).
			And().
			Term("learning").
			And().Not().
			Term("python").
			Execute()
	}
}

// BenchmarkQueryBuilder_WithBM25 benchmarks query with BM25 scoring
func BenchmarkQueryBuilder_WithBM25(b *testing.B) {
	idx := setupTestIndex()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewQueryBuilder(idx).
			Term("machine").
			And().
			Term("learning").
			ExecuteWithBM25(10)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════════

// bitmapToSlice converts a roaring bitmap to a sorted slice of ints
func bitmapToSlice(bitmap *roaring.Bitmap) []int {
	if bitmap == nil {
		return []int{}
	}

	result := make([]int, 0, bitmap.GetCardinality())
	iter := bitmap.Iterator()
	for iter.HasNext() {
		result = append(result, int(iter.Next()))
	}
	return result
}

// slicesEqual checks if two slices are equal
func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
