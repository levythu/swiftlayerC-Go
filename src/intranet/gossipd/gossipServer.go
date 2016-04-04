package gossipd

import (
    . "github.com/levythu/gurgling"
    "io/ioutil"
    . "logger"
)


func checkGossipedData(src []*GossipEntry) {

}
/*
** GOSSIP API: Posted
** Method:      POST
** URL:         [:intranet]/gossip
** Parameter:   Content(in Body): the raw body is the parameter content itself.
*/
func OnPostedGossip(req Request, res Response) {
    if ct, err:=ioutil.ReadAll(req.R().Body); err!=nil {
        Secretary.Error("gossipd::OnPostedGossip", "Fail to read data from gossiped request: "+err.Error())
        res.SendCode(500)
        return
    } else {
        var pList=ParseAll(string(ct))
        if pList==nil {
            Secretary.Error("gossipd::OnPostedGossip", "Format error for gossiped data")
            res.SendCode(403)
            return
        }


    }
}

func GetGossipRouter() Router {
    var r=ARouter()
    r.Post("/", OnPostedGossip)

    return r
}
