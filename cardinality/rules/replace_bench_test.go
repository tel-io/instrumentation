package rules_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tel-io/instrumentation/cardinality"
	"github.com/tel-io/instrumentation/cardinality/rules"
)

/*
Only equal logic

	cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
	BenchmarkEqual
	BenchmarkEqual-8   	   41247	     28634 ns/op
*/
func BenchmarkEqualDefaults(b *testing.B) {
	benchmarkEqual(b, rules.DefaultMaxRuleCount, rules.DefaultMaxSeparatorCount)
}
func BenchmarkEqualSep100(b *testing.B) {
	benchmarkEqual(b, rules.DefaultMaxRuleCount, 100)
}
func BenchmarkEqualSep1000(b *testing.B) {
	benchmarkEqual(b, rules.DefaultMaxRuleCount, 1000)
}
func BenchmarkEqualRules1000(b *testing.B) {
	benchmarkEqual(b, 1000, rules.DefaultMaxSeparatorCount)
}
func BenchmarkEqualRules1000Sep100(b *testing.B) {
	benchmarkEqual(b, 1000, 100)
}
func BenchmarkEqualRules10000(b *testing.B) {
	benchmarkEqual(b, 10000, rules.DefaultMaxSeparatorCount)
}
func BenchmarkEqualRules10000Sep1000(b *testing.B) {
	benchmarkEqual(b, 10000, 1000)
}

/*
Idle

	cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
	BenchmarkPartial
	BenchmarkPartial-8   	 2058320	       574.1 ns/op

With partial logic

	cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
	BenchmarkPartial
	BenchmarkPartial-8   	   15498	     73421 ns/op
*/
func BenchmarkPartialDefaults(b *testing.B) {
	benchmarkPartial(b, rules.DefaultMaxRuleCount, rules.DefaultMaxSeparatorCount)
}
func BenchmarkPartialSep100(b *testing.B) {
	benchmarkPartial(b, rules.DefaultMaxRuleCount, 100)
}
func BenchmarkPartialSep1000(b *testing.B) {
	benchmarkPartial(b, rules.DefaultMaxRuleCount, 1000)
}
func BenchmarkPartialRules1000(b *testing.B) {
	benchmarkPartial(b, 1000, rules.DefaultMaxSeparatorCount)
}
func BenchmarkPartialRules1000Sep100(b *testing.B) {
	benchmarkPartial(b, 1000, 100)
}
func BenchmarkPartialRules10000(b *testing.B) {
	benchmarkPartial(b, 10000, rules.DefaultMaxSeparatorCount)
}
func BenchmarkPartialRules10000Sep1000(b *testing.B) {
	benchmarkPartial(b, 10000, 1000)
}

func benchmarkEqual(b *testing.B, rulesCount, separatorsCount int) {
	var rList = make([]string, 0, rulesCount)
	var placeholder = cardinality.DefaultConfig().PlaceholderFormatter()("id")
	var separator = cardinality.DefaultConfig().PathSeparator()

	for ruleIndex := 0; ruleIndex < rulesCount; ruleIndex++ {
		var rp = make([]string, 0, rulesCount)
		repIndex := ruleIndex % separatorsCount

		for sepIndex := 0; sepIndex < separatorsCount; sepIndex++ {
			if repIndex == sepIndex {
				rp = append(rp, placeholder)
			} else {
				rp = append(rp, fmt.Sprintf("part-%d-%d", ruleIndex, sepIndex))
			}
		}

		rList = append(rList, separator+strings.Join(rp, separator))
	}

	m, errP := rules.New(rList,
		rules.WithMaxRuleCount(rulesCount+1),
		rules.WithMaxSeparatorCount(separatorsCount+1),
	)
	require.NoError(b, errP)

	url := strings.TrimSuffix(rList[len(rList)-1], placeholder) + "cardinality"

	for i := 0; i < b.N; i++ {
		m.Replace(url)
	}
}
func benchmarkPartial(b *testing.B, rulesCount, separatorsCount int) {
	var rList = make([]string, 0, rulesCount)
	var eList = make([]string, 0, rulesCount)
	var placeholder = cardinality.DefaultConfig().PlaceholderFormatter()("id")
	var separator = cardinality.DefaultConfig().PathSeparator()

	for ruleIndex := 0; ruleIndex < rulesCount; ruleIndex++ {
		repIndex := ruleIndex % separatorsCount
		var rp = make([]string, 0, rulesCount)

		for sepIndex := 0; sepIndex < separatorsCount; sepIndex++ {
			if repIndex == sepIndex {
				rp = append(rp, placeholder)
			} else {
				rp = append(rp, fmt.Sprintf("part-%d-%d", ruleIndex, sepIndex))
			}
		}

		var left, right int
		var rpl []string

		if repIndex < 2 {
			right = 2 + randInt(0, len(rp)-2)

			//fmt.Printf("A(len:%d, rep:%d) [:%d]\n", len(rp), repIndex, right)
			rpl = rp[:right]
		} else if repIndex == len(rp)-1 {
			left = randInt(0, len(rp)-2)

			//fmt.Printf("B(len:%d, rep:%d) [%d:]\n", len(rp), repIndex, left)
			rpl = rp[left:]
		} else {
			if repIndex > 1 {
				left = randInt(0, repIndex-2)
			}
			if repIndex > 1 {
				right = len(rp) - 1

				if rm := len(rp) - 1 - repIndex; rm >= 1 {
					right = randInt(0, rm-1) + repIndex + 1
				}
			}

			//fmt.Printf("C(len:%d, rep:%d) [%d:%d]\n", len(rp), repIndex, left, right)
			rpl = rp[left:right]
		}

		rList = append(rList, strings.Join(rpl, separator))
		eList = append(eList, strings.Replace(strings.Join(rp, separator), placeholder, "cardinality", 1))
	}

	m, errP := rules.New(rList,
		rules.WithMaxRuleCount(rulesCount+1),
		rules.WithMaxSeparatorCount(separatorsCount+1),
	)
	require.NoError(b, errP)

	for i := 0; i < b.N; i++ {
		m.Replace(eList[i%rulesCount])
	}
}
func randInt(min, max int) int {
	r, _ := rand.Int(rand.Reader, big.NewInt(int64(max+1-min)))
	return int(r.Int64()) + min
}
