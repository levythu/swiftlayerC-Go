package main

import (
    "runtime"
    "definition/configinfo"
    "fmt"
)

func prepEnv_SetConcurrency() {
    num:=configinfo.THREAD_UTILISED
    if (num<=0) {
        num=runtime.NumCPU()
    }
    runtime.GOMAXPROCS(num)
    fmt.Println("- Set GOMAXPROCS to ",runtime.GOMAXPROCS(-1))
}
// Only run once when start.
func startUp() {
    fmt.Println("- Swift Layer-C is starting...")
    prepEnv_SetConcurrency()
    fmt.Println("- Premise checked. Now lauching Web server...")
}
