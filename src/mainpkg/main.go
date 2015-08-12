package mainpkg

import (
    "fmt"
    "testpkg1"
    "testpkg1/pkg2"
)

func main() {
	fmt.Printf("Hello, world.\n")
    testpkg1.MbRun()
    pkg2.MbRun()
}
