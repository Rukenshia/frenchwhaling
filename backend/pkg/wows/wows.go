package wows

// EventStartTime is the Unix timestamp when the event starts
var EventStartTime = map[string]int{
	"eu":   1608768000,
	"com":  1608768000,
	"ru":   1608768000,
	"asia": 1608768000,
}

var ActiveEvent = Snowflake2020{}
