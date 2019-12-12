package wows

// EventStartTime is the Unix timestamp when the event starts
var EventStartTime = map[string]int{
	"eu":   1576141200,
	"com":  1576069200,
	"ru":   1576047600,
	"asia": 1576108800,
}

var ActiveEvent = Snowflake2019{}
