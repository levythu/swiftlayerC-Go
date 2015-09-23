package logger

import (
    . "definition"
)

var Secretary Logger=&ConsoleLogger{}

type Logger interface {
    LogD(c Tout)
    WarnD(c Tout)
    ErrorD(c Tout)
    Log(pos string, c Tout)
    Warn(pos string, c Tout)
    Error(pos string, c Tout)
}
