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
	// SuperContainer is a special container that can usually only be received by chance in
	// daily containers or through missions/events
	SuperContainer Resource = iota
	// AnniversaryCamouflages are a special camouflage
	AnniversaryCamouflages Resource = iota
	// AnniversaryContainers are a special container for the WoWS Anniversary
	AnniversaryContainers Resource = iota
	// FestiveToken is a special resource for the 2021 anniversary event
	FestiveToken = iota
	// FestiveTokenAndAnniversaryContainer is a composite of one FestiveToken and an AnniversaryContainer
	// I did this as a quick hack because the entire code relies on only one resource per redeemable
	FestiveTokenAndAnniversaryContainer = iota
)
