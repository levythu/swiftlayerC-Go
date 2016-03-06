package logger

import (
    . "definition"
)

type VoidLogger struct {
    // Nothing
}

func (_ *VoidLogger)LogD(c Tout) {
    // NOTHING
}
func (_ *VoidLogger)WarnD(c Tout) {
    // NOTHING
}
func (_ *VoidLogger)ErrorD(c Tout) {
    // NOTHING
}
func (_ *VoidLogger)Log(pos string, c Tout) {
    // NOTHING
}
func (_ *VoidLogger)Warn(pos string, c Tout) {
    // NOTHING
}
func (_ *VoidLogger)Error(pos string, c Tout) {
    // NOTHING
}
