package log

import "testing"

type testlevel struct {
	value    string
	loglevel string
	e        bool
}

var testpairs = []*testlevel{
	&testlevel{"error", "error", false},
	&testlevel{"Warn", "warning", false},
	&testlevel{"", "", true},
	&testlevel{"12", "", true},
}

func TestSetLogLevel(t *testing.T) {
	for _, pair := range testpairs {
		err := SetLogLevel(pair.value)
		if pair.e {
			if err == nil {
				t.Error("For", pair.value,
					"expected non nil error",
					"got nil error")
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
