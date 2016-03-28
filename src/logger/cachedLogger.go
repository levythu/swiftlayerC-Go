package logger

import (
    "fmt"
    . "definition"
    "time"
    "sync"
    "strconv"
)

type cachedLoggerEntry struct {
    pos string
    content string
    time time.Time
}
type CachedLogger struct {
    loggingQueue [][]cachedLoggerEntry
    queueTail []int     // Tain points to the next of the end
    length []int
    lock []sync.RWMutex
    capacity int


    doLog bool
    doWarn bool
    doErr bool

    next Logger
}

var SecretaryCache=&CachedLogger {
    doLog: true,
    doWarn: true,
    doErr: true,

    next: nil,
}

func (this *CachedLogger)Init(channels int/*=3*/, capacity int) *CachedLogger {
    this.loggingQueue=make([][]cachedLoggerEntry, channels)
    this.queueTail=make([]int, channels)
    this.length=make([]int, channels)
    this.lock=make([]sync.RWMutex, channels)
    for i:=0; i<channels; i++ {
        this.loggingQueue[i]=make([]cachedLoggerEntry, capacity)
        this.queueTail[i]=0
        this.length[i]=0
    }
    this.capacity=capacity

    return this
}

func (this *CachedLogger)RecordInChannel(channelNum int, pos string, content string, time time.Time) {
    this.lock[channelNum].Lock()
    defer this.lock[channelNum].Unlock()

    var p=this.queueTail[channelNum]
    this.loggingQueue[channelNum][p].pos=pos
    this.loggingQueue[channelNum][p].content=content
    this.loggingQueue[channelNum][p].time=time
    this.queueTail[channelNum]++

    if this.length[channelNum]<this.capacity {
        this.length[channelNum]++
    }
}

func (this *CachedLogger)LogD(c Tout) {
    if this.doLog {
        this.RecordInChannel(0, "", fmt.Sprint(c), time.Now())
    }
    if this.next!=nil {
        this.next.LogD(c)
    }
}
func (this *CachedLogger)WarnD(c Tout) {
    if this.doWarn {
        this.RecordInChannel(1, "", fmt.Sprint(c), time.Now())
    }
    if this.next!=nil {
        this.next.WarnD(c)
    }
}
func (this *CachedLogger)ErrorD(c Tout) {
    if this.doErr {
        this.RecordInChannel(2, "", fmt.Sprint(c), time.Now())
    }
    if this.next!=nil {
        this.next.ErrorD(c)
    }
}
func (this *CachedLogger)Log(pos string, c Tout) {
    if this.doLog {
        this.RecordInChannel(0, pos, fmt.Sprint(c), time.Now())
    }
    if this.next!=nil {
        this.next.Log(pos, c)
    }
}
func (this *CachedLogger)Warn(pos string, c Tout) {
    if this.doWarn {
        this.RecordInChannel(1, pos, fmt.Sprint(c), time.Now())
    }
    if this.next!=nil {
        this.next.Warn(pos, c)
    }
}
func (this *CachedLogger)Error(pos string, c Tout) {
    if this.doErr {
        this.RecordInChannel(2, pos, fmt.Sprint(c), time.Now())
    }
    if this.next!=nil {
        this.next.Error(pos, c)
    }
}

func (this *CachedLogger)SetLevel(level int) {
    this.doErr  =(level & 1!=0)
    this.doWarn =(level & 2!=0)
    this.doLog  =(level & 4!=0)

    Secretary.Log("logger.CachedLogger::SetLevel", "Log Level is set to "+strconv.Itoa(level & 7))

    if this.next!=nil {
        this.next.SetLevel(level)
    }
}
func (this *CachedLogger)Chain(obj Logger) Logger {
    if this.next==nil {
        this.next=obj
    } else {
        this.next.Chain(obj)
    }
    return this
}
