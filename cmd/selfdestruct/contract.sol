pragma solidity ^0.8.7;
contract Selfdestructer {

    mapping(uint256 => uint256) map;
    uint256 size;

    function Store() public{
        uint256 rnd = block.difficulty;
        while (gasleft() > 60000) {
            assembly {
                sstore(rnd, rnd)
                rnd := add(rnd, 1)
            }
        }
        size += (rnd - block.difficulty) * 32;
    }

    function Destruct() public {
        selfdestruct(payable(msg.sender));
    }

    function Size() public view returns (uint256) {
        return size;
    }
}
