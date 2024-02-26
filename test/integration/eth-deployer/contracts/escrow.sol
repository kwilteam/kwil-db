// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract Escrow {
  
    event Credit(address _from, uint256 _amount);

    event Test(string text);

    IERC20 public escrowToken;

    // Mapping of tokens locked in the escrow contract by each user
    mapping(address => uint256) public deposits;

    uint256 public escrowedFunds; // total amount of funds in the escrow contract



    constructor(address _escrowToken) {
        escrowToken = IERC20(_escrowToken);
    }

    /**
        * @dev Deposit funds into the escrow contract
        * @param amt The amount of tokens to deposit into the escrow contract by the caller

        * @notice This function will transfer the tokens from the caller to the escrow contract

     */
    function deposit(uint256 amt) public payable {
        require(escrowToken.transferFrom(msg.sender, address(this), amt), "Deposit failed: token did not successfully transfer");

        deposits[msg.sender] += amt;

        emit Credit(msg.sender, amt);
    }

    /**
        * Dummy function just for testing purposes
     */
    function test() public payable {
        emit Test("test string");
    }

    function balance(address wallet) public view returns(uint256) {
        return deposits[wallet];
    }
}