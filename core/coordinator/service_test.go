package coordinator

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCombinations(t *testing.T) {
	params := map[string][]interface{}{
		"fast": {10, 20},
		"slow": {50},
	}

	results := generateCombinations(params)
	assert.Len(t, results, 2)
	
	// Check combinations
	hasFast10 := false
	hasFast20 := false
	for _, res := range results {
		if res["fast"] == 10 && res["slow"] == 50 {
			hasFast10 = true
		}
		if res["fast"] == 20 && res["slow"] == 50 {
			hasFast20 = true
		}
	}
	assert.True(t, hasFast10)
	assert.True(t, hasFast20)
}

func TestGenerateCombinations_Empty(t *testing.T) {
	params := map[string][]interface{}{}
	results := generateCombinations(params)
	assert.Len(t, results, 0)
}

func TestGenerateCombinations_Single(t *testing.T) {
	params := map[string][]interface{}{
		"fast": {10},
	}
	results := generateCombinations(params)
	assert.Len(t, results, 1)
	assert.Equal(t, 10, results[0]["fast"])
}
