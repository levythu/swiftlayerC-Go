package distributedvc

// Unit test for kernel/distributedvc

import (
    "testing"
    "fmt"
    "strconv"
    "sync"
    "math/rand"
)

func _TestAutoDormant(t *testing.T) {
    var wg sync.WaitGroup
    for i:=0; i<10050; i++ {
        wg.Add(1)
        go (func(num int) {
            //fmt.Println("Thread #", num, "is running.")
            var name="name "+strconv.Itoa(num)
            var des=GetFD(name)
            des.GraspReader()
            des.Read()
            des.ReleaseReader()
            des.Release()
            wg.Done()
        })(i)
    }
    wg.Wait()
    fmt.Println(dormant.Length)
}

func TestAllRound(t *testing.T) {
    var wg sync.WaitGroup
    for i:=0; i<10050; i++ {
        wg.Add(1)
        go (func(num int) {
            //fmt.Println("Thread #", num, "is running.")
            var name="name "+strconv.Itoa(num)
            var des=GetFD(name)
            des.GraspReader()
            des.Read()
            des.ReleaseReader()
            des.Release()
            wg.Done()
        })(rand.Intn(99))
    }
    wg.Wait()
    fmt.Println(dormant.Length)
}
