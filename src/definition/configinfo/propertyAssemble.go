package configinfo

import (
    . "logger"
)

var NODE_NUMBER int
var NODE_NUMS_IN_ALL int
var AUTO_COMMIT_PER_INTRAMERGE int
var SWIFT_AUTH_URL string
var SWIFT_PROXY_URL string
var INDEX_FILE_CHECK_MD5 bool
var THREAD_UTILISED int

var KEYSTONE_USERNAME string
var KEYSTONE_TENANT string
var KEYSTONE_PASSWORD string

var MAX_NUMBER_OF_ACTIVE_FD int
var MAX_NUMBER_OF_DORMANT_FD int

func maxInt(n1, n2 int) int {
    if n1>n2 {
        return n1
    } else {
        return n2
    }
}
func InitAll() bool {
    errorAssert(AppendFileToJSON("conf/nodeinfo.json", conf), "Reading conf/nodeinfo.json")
    errorAssert(AppendFileToJSON("conf/accountinfo.debug.noupload.json", conf), "Reading conf/accountinfo.debug.noupload.json")
    errorAssert(AppendFileToJSON("conf/kernelinfo.json", conf), "Reading conf/kernelinfo.json")

    NODE_NUMBER                     =int(extractProperty("node_number").(float64))
    NODE_NUMS_IN_ALL                =int(extractProperty("node_nums_in_all").(float64))
    AUTO_COMMIT_PER_INTRAMERGE      =int(extractProperty("auto_commit_per_intramerge").(float64))
    SWIFT_AUTH_URL                  =extractProperty("swift_auth_url").(string)
    SWIFT_PROXY_URL                 =extractProperty("swift_proxy_url").(string)
    INDEX_FILE_CHECK_MD5            =extractProperty("index_file_check_md5").(bool)
    THREAD_UTILISED                 =int(extractProperty("thread_utilised").(float64))

    KEYSTONE_USERNAME               =extractProperty("keystone_username").(string)
    KEYSTONE_TENANT                 =extractProperty("keystone_tenant").(string)
    KEYSTONE_PASSWORD               =extractProperty("keystone_password").(string)

    MAX_NUMBER_OF_ACTIVE_FD         =int(extractProperty("max_number_of_active_fd").(float64))
    if MAX_NUMBER_OF_ACTIVE_FD<100 {
        Secretary.WarnD("The configuration variable MAX_NUMBER_OF_ACTIVE_FD is too small and is set to 100 automatically.")
        MAX_NUMBER_OF_ACTIVE_FD=100
    }
    MAX_NUMBER_OF_DORMANT_FD        =int(extractProperty("max_number_of_dormant_fd").(float64))
    if MAX_NUMBER_OF_DORMANT_FD<100 {
        Secretary.WarnD("The configuration variable MAX_NUMBER_OF_DORMANT_FD is too small and is set to 100 automatically.")
        MAX_NUMBER_OF_DORMANT_FD=100
    }

    return true
}

var _=InitAll()
