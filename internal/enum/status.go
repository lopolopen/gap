package enum

//go:generate go tool shoot enum -sql -gorm -type=Status

type Status int32

const (
	StatusFailed  Status = -1
	StatusPending Status = iota //1
	StatusProcessing
	StatusSucceeded
)
