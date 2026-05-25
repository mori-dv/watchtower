package checker

type Status string

const (
    StatusUp        Status = "UP"
    StatusDown      Status = "DOWN"
    StatusDegraded  Status = "DEGRADED"
)
