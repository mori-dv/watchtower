package checker

import "time"


type Result struct {
    TargetName string
    TargetType string

    Status Status

    Latency time.Duration

    Error string

    CheckedAt time.Time
}