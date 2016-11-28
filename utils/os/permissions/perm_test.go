package permissions

import (
	"os"
	"testing"
)

type testperm struct {
	value    string
	expected os.FileMode
	e        bool // whether we expect an error or not
}

var testpairs = []*testperm{
	&testperm{
		value:    "06444",
		expected: 0,
		e:        true,
	},
	&testperm{
		value:    "64",
		expected: 0,
		e:        true,
	},
	&testperm{
		value:    "0a44",
		expected: 0,
		e:        true,
	},
	&testperm{
		value:    "06a4",
		expected: 0,
		e:        true,
	},
	&testperm{
		value:    "064a",
		expected: 0,
		e:        true,
	},
	&testperm{
		value:    "0644",
		expected: os.FileMode(0644),
		e:        false,
	},
	&testperm{
		value:    "644",
		expected: os.FileMode(0644),
		e:        false,
	},
	&testperm{
		value:    "0600",
		expected: os.FileMode(0600),
		e:        false,
	},
	&testperm{
		value:    "600",
		expected: os.FileMode(0600),
		e:        false,
	},
	&testperm{
		value:    "wrong_value",
		expected: 0,
		e:        true,
	},
}

func TestStringToFileMode(t *testing.T) {
	for _, pair := range testpairs {
		filemode, err := StringToFileMode(pair.value)
		if pair.e {
			if err == nil {
				t.Error("For", pair.value,
					"expected non nil error",
					"got nil error")
			}
		} else {
			if err != nil {
				t.Error("For", pair.value,
					"expected nil error",
					"got non nil error")
			}
			if pair.expected != filemode {
				t.Error("For", pair.value,
					"expected", pair.expected,
					"got", filemode)
			}
		}
	}
}
