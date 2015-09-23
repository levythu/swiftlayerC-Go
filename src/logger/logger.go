package logger

import (
    . "definition"
)

var Secretary Logger=&ConsoleLogger{}

type Logger interface {
    func LogD(c Tout)
    func WarnD(c Tout)
    func ErrorD(c Tout)
    func Log(pos string, c Tout)
    func Warn(pos string, c Tout)
    func Error(pos string, c Tout)
}
