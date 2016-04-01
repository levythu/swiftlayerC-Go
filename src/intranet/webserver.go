package intranet

// used for administrative maintainance and inter-communication between servers as well

import (
    . "github.com/levythu/gurgling"
    conf "definition/configinfo"
    . "logger"
)

func Entry(exit chan bool) {
    defer (func(){
        exit<-false
    })()

    var rootRouter=ARouter()

    rootRouter.Get("/", func(res Response) {
        res.Redirect("/admin")
    })
    if r:=getGossipRouter(); r!=nil {
        rootRouter.Use("/gossip", r)
    }
    if r:=getPingRouter(); r!=nil {
        rootRouter.Use("/ping", r)
    }
    if r:=getAdminPageRouter(); r!=nil {
        rootRouter.Use("/admin", r)
    }

    Secretary.Log("intranet::Entry()", "Now launching intranet service at "+conf.INNER_SERVICE_LISTENER)
    var err=rootRouter.Launch(conf.INNER_SERVICE_LISTENER)
    if err!=nil {
        Secretary.Error("intranet::Entry()", "HTTP Server terminated: "+err.Error())
    }
}
