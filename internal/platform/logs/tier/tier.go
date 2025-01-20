package tier

type Tier string

const (
	Archive     Tier = "archive"
	Frequent    Tier = "frequent_search"
	Unspecified Tier = "unspecified"
)

const (
	LimitFrequent int = 12000
	LimitArchive  int = 50000
)
