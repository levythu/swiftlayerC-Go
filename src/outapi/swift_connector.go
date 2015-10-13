package outapi

// Used for providing a united connector.

import (
    . "definition/configinfo"
)

var DefaultConnector=ConnectbyAuth(
    GetProperty_Node("keystone_username").(string),
    GetProperty_Node("keystone_password").(string),
    GetProperty_Node("keystone_tenant").(string))
