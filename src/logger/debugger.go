package logger

import (
    . "definition"
)

var Insider Dubugger=&consoleDebugger{}

type Dubugger interface {
    LogD(c Tout)
    Log(pos string, c Tout)
}
