package log

import "testing"

type testpair struct {
	value    string
	loglevel string
	e        bool
}

var testpairs = []*testpair{
	&testpair{"error", "error", false},
	&testpair{"Warn", "warning", false},
	&testpair{"", "", true},
	&testpair{"12", "", true},
}

func TestSetLogLevel(t *testing.T) {
	for _, pair := range testpairs {
		err := SetLogLevel(pair.value)
		if pair.e {
			if err == nil {
				t.Error("For", pair.value,
					"expected non nil error",
					"got", "nil error")
			}
		} else {
			resultLogLevel := Msg.Level.String()
			if resultLogLevel != pair.loglevel {
				t.Error("For", pair.value,
					"expected", pair.loglevel,
					"got", resultLogLevel)
			}
		}
	}
}
