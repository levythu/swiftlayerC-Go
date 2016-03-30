package public

import (
    . "logger"
)

func Terminated() {
    Secretary.Log("mainpkg::Terminated", "Midware-MH2 is about to terminate...")
}
