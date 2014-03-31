package internaltypes

type FlagType uint

const (
	FlagTypeNone FlagType = iota
	FlagTypeFocused
	FlagTypePending
)
