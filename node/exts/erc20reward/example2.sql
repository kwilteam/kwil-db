

USE erc20_rewards {
    chain: '',
} AS usdc;

USE erc20_rewards {

} AS weth;

-- sync ...

CREATE ACTION do_something() public {
    usdc.issue('0x', 123);
};

UNUSE usdc; -- funds get lost