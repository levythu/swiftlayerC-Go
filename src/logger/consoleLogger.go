package logger

import (
    "fmt"
    . "definition"
    "time"
)

type ConsoleLogger struct {
    // Nothing
}

func (_ *ConsoleLogger)LogD(c Tout) {
    fmt.Println("<Log  @", time.Now().Format(time.RFC3339)+">", c)
}
func (_ *ConsoleLogger)WarnD(c Tout) {
    fmt.Println("<Warn @", time.Now().Format(time.RFC3339)+">", c)
}
func (_ *ConsoleLogger)ErrorD(c Tout) {
    fmt.Println("<Err  @", time.Now().Format(time.RFC3339)+">", c)
}
func (_ *ConsoleLogger)Log(pos string, c Tout) {
    fmt.Println("<Log  @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
}
func (_ *ConsoleLogger)Warn(pos string, c Tout) {
    fmt.Println("<Warn @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
}
func (_ *ConsoleLogger)Error(pos string, c Tout) {
    fmt.Println("<Err  @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
}
