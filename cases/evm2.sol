pragma solidity ^0.5.0;

contract EVMTest2 {

    function simpleRequire() public returns (bool) {
        require(0 > 1);
        return true;
    }
}
