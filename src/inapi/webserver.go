package inapi

import (
    . "github.com/levythu/gurgling"
    "inapi/containermanage"
    "inapi/fsmanage"
    "inapi/streamio"
)

func Entry() {
    var rootRouter=ARouter()
    rootRouter.Use("/fs", fsmanage.FMRouter())
    rootRouter.Use("/io", streamio.IORouter())
    rootRouter.Use("/cn", containermanage.CMRouter())

    rootRouter.Launch(":9144")
}
