package inapi

import (
    . "github.com/levythu/gurgling"
    "inapi/containermanage"
)

func Entry() {
    var rootRouter=ARouter()
    rootRouter.Use("/container", containermanage.CMRouter())

    rootRouter.Launch(":9144")
}
