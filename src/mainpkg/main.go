package main

import (
    "inapi"
    . "mainpkg/public"
    "intranet"
    . "logger"
)

func _no_use_() {
    inapi.Entry(nil)
}

func main() {
    StartUp()

    var exitCh=make(chan bool)

    go intranet.Entry(exitCh)
    go inapi.Entry(exitCh)
    go WaitForSig(exitCh)

    _=<-exitCh
    Secretary.Log("mainpkg::main", "Midware-MH2 is about to terminate...")

}
