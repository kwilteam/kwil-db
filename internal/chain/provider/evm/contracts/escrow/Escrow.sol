// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract Escrow {
    event Deposit(
        address caller,
        uint256 amount
    );

    event Withdrawal(
        address receiver,
        address indexed caller,
        uint256 amount,
        uint256 fee,
        string nonce
    );

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

        emit Deposit(msg.sender, amt);
    }

    // function returnDeposit(address recipient, uint256 amt, uint256 fee, string memory nonce) public {
    //     uint256 total = amt + fee;
    //     require(pools[msg.sender][recipient] >= total, "Not enough to transfer back");
    //     require(escrowToken.transfer(recipient, amt), "Could not transfer funds back to owner");
    //     if (fee > 0) {
    //         require(escrowToken.transfer(msg.sender, fee), "Could not transfer funds to validator");
    //     }

    //     pools[msg.sender][recipient] -= total;

    //     emit Withdrawal(recipient, msg.sender, amt, fee, nonce);
    // }

    function balance(address wallet) public view returns(uint256) {
        return deposits[wallet];
    }
}