pragma solidity ^0.5.0;

contract EVMTest {

    event transferEvent(address from, address to, uint256 value);

    function nativeTransfer(address _to, uint _value) public returns (bool) {
        address _toContract = 0x0000000000000000000000000000000000000103;

        bool succeed;
        bytes memory returnData;
        (succeed, returnData) = _toContract.call(abi.encodePacked(bytes4(keccak256(abi.encodePacked("transfer", "(address,uint256)"))), abi.encode(_to, _value)));
        require(succeed, "native transfer failed");

        emit transferEvent(msg.sender, _to, _value);
        return true;
    }
}
