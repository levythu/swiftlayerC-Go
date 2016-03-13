package logger

import (
    . "definition"
)

type voidLogger struct {
    // Nothing
}

func (_ *voidLogger)LogD(c Tout) {
    // NOTHING
}
func (_ *voidLogger)WarnD(c Tout) {
    // NOTHING
}
func (_ *voidLogger)ErrorD(c Tout) {
    // NOTHING
}
func (_ *voidLogger)Log(pos string, c Tout) {
    // NOTHING
}
func (_ *voidLogger)Warn(pos string, c Tout) {
    // NOTHING
}
func (_ *voidLogger)Error(pos string, c Tout) {
    // NOTHING
}

func (_ *voidLogger)SetLevel(level int) {
    // NOTHING
}
