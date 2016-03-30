package main

import (
    "inapi"
    . "mainpkg/public"
    "intranet"
)

func _no_use_() {
    inapi.Entry()
}

func main() {
    StartUp()

    go intranet.Entry()
    inapi.Entry()

}
