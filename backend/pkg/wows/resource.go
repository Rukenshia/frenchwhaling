package wows

type Resource uint

const (
	// RepublicTokens are tokens that can be earned during the update 0.8.6
	RepublicTokens Resource = iota
	// Coal is a universal resource in the WoWS Armory
	Coal Resource = iota
)
