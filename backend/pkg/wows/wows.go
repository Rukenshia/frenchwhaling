package wows

// EventStartTime is the Unix timestamp when the event starts
var EventStartTime = map[string]int{
	"eu":   1576137600,
	"com":  1576137600,
	"ru":   1576137600,
	"asia": 1576137600,
}

var ActiveEvent = Snowflake2019{}
