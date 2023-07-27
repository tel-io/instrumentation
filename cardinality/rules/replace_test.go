package rules_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tel-io/instrumentation/cardinality"
	"github.com/tel-io/instrumentation/cardinality/rules"
)

func TestPartial(t *testing.T) {
	gP, errP := rules.New([]string{
		"a20/:XX",
		":XX/b21",
		"b22/:XX",
		":XX/c23",

		"b26/:XX/:YY",
		":XX/:YY/c27",
		":XX/:YY/d28",
		"a29/:XX/:YY",

		"b30/:XX",
		":XX/c31",
	})
	assert.NoError(t, errP)

	tests := map[string]string{
		"/a20/b20/c20": "/a20/:XX/c20",
		"/a21/b21/c21": "/:XX/b21/c21",
		"/a22/b22/c22": "/a22/b22/:XX",
		"/a23/b23/c23": "/a23/:XX/c23",

		"/a26/b26/c26/d26": "/a26/b26/:XX/:YY",
		"/a27/b27/c27/d27": "/:XX/:YY/c27/d27",
		"/a28/b28/c28/d28": "/a28/:XX/:YY/d28",
		"/a29/b29/c29/d29": "/a29/:XX/:YY/d29",
		"/a30/b30/c30/d30": "/a30/b30/:XX/d30",
		"/a31/b31/c31/d31": "/a31/:XX/c31/d31",
	}

	var list = cardinality.ReplacerList{gP}
	for url, exp := range tests {
		assert.Equal(t, exp, list.Apply(url))
	}
}

func TestEqual(t *testing.T) {
	gP, errP := rules.New([]string{
		"/:AA/b01/c01/:DD",
		"/a02/:BB/c02/:DD",
		"/a03/b03/:CC/:DD",
		"/a04/b04/c04/:DD",
		"/:AA/b05/c05/d05",
		"/:AA/:BB/c06/d06",
		"/:AA/b07/:CC/d07",
		"/:AA/b08/c08/:DD",
		"/a09/:BB/:CC/d09",
		"/a10/:BB/:CC/:DD",
		"/:AA/:BB/:CC/d11",
	})
	assert.NoError(t, errP)

	tests := map[string]string{
		"/a01/b01/c01/d01": "/:AA/b01/c01/:DD",
		"/a02/b02/c02/d02": "/a02/:BB/c02/:DD",
		"/a03/b03/c03/d03": "/a03/b03/:CC/:DD",
		"/a04/b04/c04/d04": "/a04/b04/c04/:DD",
		"/a05/b05/c05/d05": "/:AA/b05/c05/d05",
		"/a06/b06/c06/d06": "/:AA/:BB/c06/d06",
		"/a07/b07/c07/d07": "/:AA/b07/:CC/d07",
		"/a08/b08/c08/d08": "/:AA/b08/c08/:DD",
		"/a09/b09/c09/d09": "/a09/:BB/:CC/d09",
		"/a10/b10/c10/d10": "/a10/:BB/:CC/:DD",
		"/a11/b11/c11/d11": "/:AA/:BB/:CC/d11",

		"/a12/b12/c12/d12": "/a12/b12/c12/d12",
		"/a13/b13/c13":     "/a13/b13/c13",
		"/a14/b14":         "/a14/b14",
		"/a15":             "/a15",

		"/a16/b16/c16/d16": "/a16/b16/c16/d16",
		"/a17/b17/c17":     "/a17/b17/c17",
		"/a18/b18":         "/a18/b18",
		"/a19":             "/a19",
	}

	var list = cardinality.ReplacerList{gP}
	for url, exp := range tests {
		assert.Equal(t, exp, list.Apply(url))
	}
}

func TestInvalidRules(t *testing.T) {
	var errP error

	_, errP = rules.New(nil)
	assert.NoError(t, errP)

	_, errP = rules.New([]string{})
	assert.NoError(t, errP)

	_, errP = rules.New([]string{""})
	assert.Error(t, errP)

	_, errP = rules.New([]string{"/"})
	assert.Error(t, errP)

	_, errP = rules.New([]string{"//"})
	assert.Error(t, errP)

	_, errP = rules.New([]string{"/main"})
	assert.Error(t, errP)

	_, errP = rules.New([]string{strings.Repeat("/:x", rules.DefaultMaxSeparatorCount)})
	assert.Error(t, errP)

	_, err := rules.New(make([]string, rules.DefaultMaxRuleCount))
	assert.Error(t, err)
}

func TestRulesByLen1(t *testing.T) {
	gP, errP := rules.New([]string{
		"/:AA/:BB/:CC/:DD",
		"/:AA/:BB/:CC",
		"/:AA/:BB",
		"/:AA",
	})
	assert.NoError(t, errP)

	tests := map[string]string{
		"/a12/b12/c12/d12": "/:AA/:BB/:CC/:DD",
		"/a13/b13/c13":     "/:AA/:BB/:CC",
		"/a14/b14":         "/:AA/:BB",
		"/a15":             "/:AA",
	}

	var list = cardinality.ReplacerList{gP}
	for url, exp := range tests {
		assert.Equal(t, exp, list.Apply(url))
	}
}

func TestRulesByLen2(t *testing.T) {
	gP, errP := rules.New([]string{
		"/a10/:BB/c10",
		"/a11/b11/:CC",
		"/a12/:BB",
		"/:AA/b13/c13",
		"/:AA/b14",
		"/:AA",
	})
	assert.NoError(t, errP)

	tests := map[string]string{
		"/a10/b10/c10": "/a10/:BB/c10",
		"/a11/b11/c11": "/a11/b11/:CC",
		"/a12/b12":     "/a12/:BB",
		"/a13/b13/c13": "/:AA/b13/c13",
		"/a14/b14":     "/:AA/b14",
		"/a15":         "/:AA",
	}

	var list = cardinality.ReplacerList{gP}
	for url, exp := range tests {
		assert.Equal(t, exp, list.Apply(url))
	}
}

func TestBreakRule(t *testing.T) {
	gP, errP := rules.New([]string{
		"/:AA/b01/:DD/e01",
	})
	assert.NoError(t, errP)

	tests := map[string]string{
		"/a01/b01/c01/d01": "/a01/b01/c01/d01",
	}

	var list = cardinality.ReplacerList{gP}
	for url, exp := range tests {
		assert.Equal(t, exp, list.Apply(url))
	}
}

/*
Only equal logic

	cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
	BenchmarkEqual
	BenchmarkEqual-8   	   41247	     28634 ns/op
*/
func BenchmarkEqual(b *testing.B) {
	const (
		rulesCount      = rules.DefaultMaxRuleCount - 1
		separatorsCount = rules.DefaultMaxSeparatorCount - 1
	)
	var rList = make([]string, 0, rulesCount)
	var placeholder = cardinality.PlaceholderFormatter("id")

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

		rList = append(rList, rules.DefaultPathSeparator+strings.Join(rp, rules.DefaultPathSeparator))
	}

	m, errP := rules.New(rList)
	require.NoError(b, errP)

	url := strings.TrimSuffix(rList[len(rList)-1], placeholder) + "cardinality"

	for i := 0; i < b.N; i++ {
		m.Replace(url)
	}
}

/*
Without partial logic (idle)

	cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
	BenchmarkPartial
	BenchmarkPartial-8   	 2058320	       574.1 ns/op

With partial logic

	cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
	BenchmarkPartial
	BenchmarkPartial-8   	   15498	     73421 ns/op
*/
func BenchmarkPartial(b *testing.B) {
	var (
		rulesCount      = rules.DefaultMaxRuleCount - 1
		separatorsCount = rules.DefaultMaxSeparatorCount - 1
	)
	var rList = make([]string, 0, rulesCount)
	var eList = make([]string, 0, rulesCount)
	var placeholder = cardinality.PlaceholderFormatter("id")

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

		rList = append(rList, strings.Join(rpl, rules.DefaultPathSeparator))
		eList = append(eList, strings.Replace(strings.Join(rp, rules.DefaultPathSeparator), placeholder, "cardinality", 1))
	}

	m, errP := rules.New(rList)
	require.NoError(b, errP)

	for i := 0; i < b.N; i++ {
		m.Replace(eList[i%rulesCount])
	}
}

func randInt(min, max int) int {
	r, _ := rand.Int(rand.Reader, big.NewInt(int64(max+1-min)))
	return int(r.Int64()) + min
}
