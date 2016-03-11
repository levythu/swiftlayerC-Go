package logger

import (
    "fmt"
    . "definition"
    "strconv"
    "time"
)

type ConsoleLogger struct {
    doLog bool
    doWarn bool
    doErr bool
}

func (this *ConsoleLogger)LogD(c Tout) {
    if this.doLog {
        fmt.Println("<Log  @", time.Now().Format(time.RFC3339)+">", c)
    }
}
func (this *ConsoleLogger)WarnD(c Tout) {
    if this.doWarn {
        fmt.Println("<Warn @", time.Now().Format(time.RFC3339)+">", c)
    }
}
func (this *ConsoleLogger)ErrorD(c Tout) {
    if this.doErr {
        fmt.Println("<Err  @", time.Now().Format(time.RFC3339)+">", c)
    }
}
func (this *ConsoleLogger)Log(pos string, c Tout) {
    if this.doLog {
        fmt.Println("<Log  @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
    }
}
func (this *ConsoleLogger)Warn(pos string, c Tout) {
    if this.doWarn {
        fmt.Println("<Warn @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
    }
}
func (this *ConsoleLogger)Error(pos string, c Tout) {
    if this.doErr {
        fmt.Println("<Err  @", time.Now().Format(time.RFC3339)+", "+pos+">", c)
    }
}

func (this *ConsoleLogger)SetLevel(level int) {
    this.doErr  =(level & 1!=0)
    this.doWarn =(level & 2!=0)
    this.doLog  =(level & 4!=0)

    this.Log("logger.ConsoleLogger::SetLevel", "Log Level is set to "+strconv.Itoa(level & 7))
}
