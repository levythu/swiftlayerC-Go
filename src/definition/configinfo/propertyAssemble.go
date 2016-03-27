package configinfo

import (
    . "logger"
    "strconv"
)

var LOG_LEVEL int

var NODE_NUMBER int
var NODE_NUMS_IN_ALL int
var AUTO_COMMIT_PER_INTRAMERGE int
var SWIFT_AUTH_URL string
var SWIFT_PROXY_URL string
var OUTER_SERVICE_LISTENER string
var INNER_SERVICE_LISTENER string


var INDEX_FILE_CHECK_MD5 bool
var THREAD_UTILISED int

var KEYSTONE_USERNAME string
var KEYSTONE_TENANT string
var KEYSTONE_PASSWORD string

var MAX_NUMBER_OF_CACHED_ACTIVE_FD int
var MAX_NUMBER_OF_CACHED_DORMANT_FD int
var MAX_NUMBER_OF_TOTAL_ACTIVE_FD int
var MAX_NUMBER_OF_TOTAL_DORMANT_FD int

var SINGLE_FILE_SYNC_INTERVAL_MIN int64

var AUTO_MERGER_TASK_QUEUE_CAPACITY int
var MAX_MERGING_WORKER int
var REST_INTERVAL_OF_WORKER_IN_MS int
var AUTO_MERGER_DEAMON_PERIOD int

var TRIAL_INTERVAL_IN_UNREVOCABLE_IOERROR int

var ADMIN_USER string
var ADMIN_PASSWORD string

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


    LOG_LEVEL                       =int(extractProperty("log_level").(float64))
    if LOG_LEVEL>7 || LOG_LEVEL<0 {
        Secretary.WarnD("The configuration variable LOG_LEVEL is out of range and is set to 7(111b) automatically.")
        LOG_LEVEL=7
    }
    Secretary.SetLevel(LOG_LEVEL)


    NODE_NUMBER                     =int(extractProperty("node_number").(float64))
    NODE_NUMS_IN_ALL                =int(extractProperty("node_nums_in_all").(float64))
    if NODE_NUMBER>=NODE_NUMS_IN_ALL || NODE_NUMBER<0 {
        Secretary.ErrorD("The configuration variable NODE_NUMBER is supposed to be in the range [0, "+strconv.Itoa(NODE_NUMS_IN_ALL-1)+"]. There "+
            "may be some severe errors in the configuration files. Hence, please overhaul them.")
        Secretary.ErrorD("S-H2 has stopped due to servere error.")
        panic("EXIT DUE TO ASSERTION FAILURE.")
    }
    AUTO_COMMIT_PER_INTRAMERGE      =int(extractProperty("auto_commit_per_intramerge").(float64))
    if AUTO_COMMIT_PER_INTRAMERGE<1 {
        Secretary.WarnD("The configuration variable AUTO_COMMIT_PER_INTRAMERGE is too small and is set to 1 automatically.")
        AUTO_COMMIT_PER_INTRAMERGE=1
    }
    SWIFT_AUTH_URL                  =extractProperty("swift_auth_url").(string)
    SWIFT_PROXY_URL                 =extractProperty("swift_proxy_url").(string)
    OUTER_SERVICE_LISTENER          =extractProperty("outer_service_listener").(string)
    INNER_SERVICE_LISTENER          =extractProperty("inner_service_listener").(string)


    INDEX_FILE_CHECK_MD5            =extractProperty("index_file_check_md5").(bool)
    THREAD_UTILISED                 =int(extractProperty("thread_utilised").(float64))



    KEYSTONE_USERNAME               =extractProperty("keystone_username").(string)
    KEYSTONE_TENANT                 =extractProperty("keystone_tenant").(string)
    KEYSTONE_PASSWORD               =extractProperty("keystone_password").(string)



    MAX_NUMBER_OF_CACHED_ACTIVE_FD  =int(extractProperty("max_number_of_cached_active_fd").(float64))
    if MAX_NUMBER_OF_CACHED_ACTIVE_FD<100 {
        Secretary.WarnD("The configuration variable MAX_NUMBER_OF_CACHED_ACTIVE_FD is too small and is set to 100 automatically.")
        MAX_NUMBER_OF_CACHED_ACTIVE_FD=100
    }
    MAX_NUMBER_OF_CACHED_DORMANT_FD =int(extractProperty("max_number_of_cached_active_fd").(float64))
    if MAX_NUMBER_OF_CACHED_DORMANT_FD<100 {
        Secretary.WarnD("The configuration variable MAX_NUMBER_OF_CACHED_DORMANT_FD is too small and is set to 100 automatically.")
        MAX_NUMBER_OF_CACHED_DORMANT_FD=100
    }
    MAX_NUMBER_OF_TOTAL_ACTIVE_FD   =int(extractProperty("max_number_of_total_active_fd").(float64))
    if MAX_NUMBER_OF_TOTAL_ACTIVE_FD<=MAX_NUMBER_OF_CACHED_ACTIVE_FD {
        Secretary.WarnD("The configuration variable MAX_NUMBER_OF_TOTAL_ACTIVE_FD cannot be smaller than MAX_NUMBER_OF_CACHED_ACTIVE_FD. "+
            "It is set to twice the value of MAX_NUMBER_OF_CACHED_ACTIVE_FD.")
        MAX_NUMBER_OF_TOTAL_ACTIVE_FD=2*MAX_NUMBER_OF_CACHED_ACTIVE_FD
    }
    MAX_NUMBER_OF_TOTAL_DORMANT_FD  =int(extractProperty("max_number_of_total_dormant_fd").(float64))
    if MAX_NUMBER_OF_TOTAL_DORMANT_FD<=MAX_NUMBER_OF_CACHED_DORMANT_FD {
        Secretary.WarnD("The configuration variable MAX_NUMBER_OF_TOTAL_DORMANT_FD cannot be smaller than MAX_NUMBER_OF_CACHED_DORMANT_FD. "+
            "It is set to twice the value of MAX_NUMBER_OF_CACHED_DORMANT_FD.")
        MAX_NUMBER_OF_TOTAL_DORMANT_FD=2*MAX_NUMBER_OF_CACHED_DORMANT_FD
    }



    SINGLE_FILE_SYNC_INTERVAL_MIN   =int64(extractProperty("single_file_sync_interval_min_in_second").(float64))
    if SINGLE_FILE_SYNC_INTERVAL_MIN<0 {
        Secretary.WarnD("The configuration variable SINGLE_FILE_SYNC_INTERVAL_MIN cannot be negative. It is set to 0.")
        SINGLE_FILE_SYNC_INTERVAL_MIN=0
    }



    AUTO_MERGER_TASK_QUEUE_CAPACITY =int(extractProperty("auto_merger_task_queue_capacity").(float64))
    if AUTO_MERGER_TASK_QUEUE_CAPACITY<100 {
        Secretary.WarnD("The configuration variable AUTO_MERGER_TASK_QUEUE_CAPACITY is too small and is set to 100 automatically.")
        AUTO_MERGER_TASK_QUEUE_CAPACITY=100
    }
    MAX_MERGING_WORKER              =int(extractProperty("max_merging_worker").(float64))
    if MAX_MERGING_WORKER<0 {
        Secretary.WarnD("The configuration variable MAX_MERGING_WORKER cannot be negative. It is set to 0.")
        MAX_MERGING_WORKER=0
    }
    REST_INTERVAL_OF_WORKER_IN_MS   =int(extractProperty("rest_interval_of_worker_in_ms").(float64))
    if REST_INTERVAL_OF_WORKER_IN_MS<0 {
        Secretary.WarnD("The configuration variable REST_INTERVAL_OF_WORKER_IN_MS cannot be negative. It is set to 0.")
        REST_INTERVAL_OF_WORKER_IN_MS=0
    }
    AUTO_MERGER_DEAMON_PERIOD       =int(extractProperty("auto_merger_deamon_period_in_seconds").(float64))


    TRIAL_INTERVAL_IN_UNREVOCABLE_IOERROR   =int(extractProperty("trial_interval_in_unrevocable_io_error_in_ms").(float64))
    if TRIAL_INTERVAL_IN_UNREVOCABLE_IOERROR<0 {
        TRIAL_INTERVAL_IN_UNREVOCABLE_IOERROR=0
    }


    ADMIN_USER                      =extractProperty("inner_service_admin_user").(string)
    ADMIN_PASSWORD                  =extractProperty("inner_service_admin_password").(string)
    ADMIN_REFRESH_FREQUENCY         =int(extractProperty("inner_service_admin_refresh_frequency_in_second").(float64))
    if ADMIN_REFRESH_FREQUENCY<0 {
        ADMIN_REFRESH_FREQUENCY=0
    }

    return true
}

var _=InitAll()
