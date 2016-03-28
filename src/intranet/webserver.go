package intranet

// used for administrative maintainance and inter-communication between servers as well

import (
    . "github.com/levythu/gurgling"
    conf "definition/configinfo"
    . "logger"
)

func Entry() {
    var rootRouter=ARouter()

    rootRouter.Get("/", func(res Response) {
        res.Redirect("/admin")
    })
    if r:=getAdminPageRouter(); r!=nil {
        rootRouter.Use("/admin", r)
    }

    Secretary.Log("intranet::Entry()", "Now launching intranet service at "+conf.INNER_SERVICE_LISTENER)
    var err=rootRouter.Launch(conf.INNER_SERVICE_LISTENER)
    if err!=nil {
        Secretary.Error("intranet::Entry()", "HTTP Server terminated: "+err.Error())
    }
}
