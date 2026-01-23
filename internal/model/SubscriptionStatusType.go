package model

type SubscriptionStatusType int

const (
	INACTIVESUB SubscriptionStatusType = iota
	INCOMPLETE
	ACTIVESUB
	PASTDUE
	CANCELED
)

var SubscriptionStatusTypeLabel = map[SubscriptionStatusType]string{
	INACTIVESUB: "Inactive",
	INCOMPLETE:  "Incomplete",
	ACTIVESUB:   "Active",
	PASTDUE:     "Past Due",
	CANCELED:    "Canceled",
}

func (sub SubscriptionStatusType) String() string {
	return SubscriptionStatusTypeLabel[sub]
}
