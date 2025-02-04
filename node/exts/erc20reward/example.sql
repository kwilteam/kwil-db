kwil_erc20_rewards.prepare('0xabc', 'eth', '1d')
kwil_erc20_rewards.prepare('0xdef', 'base', '1d') -- id of def
-- register and begin syncing info

CREATE ACTION do_something() public {
    $id_for_smart_contract_1 := kwil_erc20_rewards.id('0xabc', 'eth'); -- id of abc
    $id_for_smart_contract_2 := kwil_erc20_rewards.id('0xdef', 'base'); -- id of abc

    $reward := 1 + 2 + 0.5;
    -- numeric(78,0) numeric(78,24)
    kwil_erc20_rewards.issue($id, '0x', $reward::text); -- numeric(102, 24) OR text
};

kwil_erc20_rewards.disable($id); -- keeps data, but doesn't issue