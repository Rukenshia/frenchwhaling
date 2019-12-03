package wows

type Resource uint

const (
	// RepublicTokens are tokens that can be earned during the update 0.8.6
	RepublicTokens Resource = iota
	// Coal is a universal resource in the WoWS Armory
	Coal Resource = iota
	// Steel is a harder to receive resource, used in the WoWS Armory for ships
	Steel Resource = iota
	// SantaGiftContainer is a special container for the Snowflake 2019 event (0.8.11)
	SantaGiftContainer Resource = iota
)
