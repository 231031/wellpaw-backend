package model

type PaymentStatusType int

const (
	PENDING PaymentStatusType = iota
	SUCCEEDED
	FAILED
)

var PaymentStatusTypeLabel = map[PaymentStatusType]string{
	PENDING:   "Pending",
	SUCCEEDED: "Succeeded",
	FAILED:    "Failed",
}

func (payment PaymentStatusType) String() string {
	return PaymentStatusTypeLabel[payment]
}
