package headers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	headers := FromMap(map[string][]string{
		"Hello": {"world"},
		"Some":  {"multiple", "values"},
	})

	t.Run("ValueOr_Existing", func(t *testing.T) {
		value := headers.ValueOr("Some", "this should not happen")
		require.Equal(t, "multiple", value)
	})

	t.Run("ValueOr_NonExisting", func(t *testing.T) {
		value := headers.ValueOr("Random", "this SHOULD happen")
		require.Equal(t, "this SHOULD happen", value)
	})

	t.Run("Value", func(t *testing.T) {
		value := headers.Value("Random")
		require.Empty(t, value)
	})

	t.Run("Values_Existing", func(t *testing.T) {
		headers := FromMap(map[string][]string{
			"Hello": {"world"},
			"Some":  {"multiple", "values"},
		})

		headers.Add("Hello", "nether")
		headers.Add("Some", "injustice")

		values := headers.Values("Some")
		require.Equal(t, []string{"multiple", "values", "injustice"}, values)
	})

	t.Run("Values_NonExisting", func(t *testing.T) {
		values := headers.Values("Random")
		require.Empty(t, values)
	})

	t.Run("Has_Existing", func(t *testing.T) {
		require.True(t, headers.Has("Hello"))
	})

	t.Run("Has_Existing", func(t *testing.T) {
		require.False(t, headers.Has("Random"))
	})

	t.Run("Add", func(t *testing.T) {
		headers := FromMap(map[string][]string{
			"Hello": {"world"},
			"Some":  {"multiple", "values"},
		})

		headers.Add("SomeHeader", "SomeValue")
		headers.Add("Hello", "Pavlo")
		require.Equal(t, []string{"SomeValue"}, headers.Values("SomeHeader"))
		require.Equal(t, []string{"multiple", "values"}, headers.Values("Some"))
		require.Equal(t, []string{"world", "Pavlo"}, headers.Values("Hello"))
	})

	t.Run("Keys", func(t *testing.T) {
		// using direct slice, as determined positions of keys are required
		headers := Headers{
			headers: []string{"Hello", "World", "Some", "multiple", "Some", "values"},
		}
		headers.Add("Hello", "nether")
		require.Equal(t, []string{"Hello", "Some"}, headers.Keys())
	})

	t.Run("Values", func(t *testing.T) {
		headers := FromMap(map[string][]string{
			"Hello": {"world"},
			"Some":  {"multiple", "values"},
		})

		headers.Add("Hello", "nether")
		headers.Add("Some", "injustice")
		require.Equal(t, []string{"multiple", "values", "injustice"}, headers.Values("some"))
	})
}
