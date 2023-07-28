package cardinality_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tel-io/instrumentation/cardinality/auto"
	"github.com/tel-io/instrumentation/cardinality/rules"
)

var url = "/player/update/123/file/favicon.ico/550e8400-e29b-41d4-a716-446655440000"
var result = "/player/update/{id}/file/{resource}/{uuid}"

func BenchmarkAuto(b *testing.B) {
	r := auto.NewHttp()
	for i := 0; i < b.N; i++ {
		require.Equal(b, result, r.Replace(url))
	}
}
func BenchmarkRules(b *testing.B) {
	r, err := rules.New([]string{result})
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		require.Equal(b, result, r.Replace(url))
	}
}
