package outapi

// Used for providing a united connector.

import (
    . "definition/configinfo"
)

var DefaultConnector=ConnectbyAuth(
    GetProperty_Node("keystone_username"),
    GetProperty_Node("keystone_password"),
    GetProperty_Node("keystone_tenant"))
