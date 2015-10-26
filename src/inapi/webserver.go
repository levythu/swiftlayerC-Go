package inapi

import (
    . "github.com/levythu/gurgling"
    "inapi/containermanage"
    "inapi/fsmanage"
)

func Entry() {
    var rootRouter=ARouter()
    rootRouter.Use("/container", containermanage.CMRouter())
    rootRouter.Use("/fs", fsmanage.FMRouter())

    rootRouter.Launch(":9144")
}
