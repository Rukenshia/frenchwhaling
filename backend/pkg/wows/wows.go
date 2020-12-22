package wows

// EventStartTime is the Unix timestamp when the event starts
var EventStartTime = map[string]int{
	"eu":   1608613200,
	"com":  1608544800,
	"ru":   1608516000,
	"asia": 1608584400,
}

var ActiveEvent = Snowflake2020{}
