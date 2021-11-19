package wows

// EventStartTime is the Unix timestamp when the event starts
var EventStartTime = map[string]int{
	"eu":   1637215200,
	"com":  1637150400,
	"ru":   1637128800,
	"asia": 1637179200,
}

var ActiveEvent = Snowflake2021{}
