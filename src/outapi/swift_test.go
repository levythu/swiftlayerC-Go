package outapi

import (
    "testing"
)

func TestContainerCreation(t *testing.T) {
    var io=NewSwiftio(DefaultConnector, "testcon02")
    io.test_Container()
}
