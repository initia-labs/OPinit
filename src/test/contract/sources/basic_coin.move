module addr::basic_coin {
    // Import the std::coin, std::string, and std::signer modules for use in this module.
    use std::coin;
    use std::string;
    use std::signer;

    // Define the Coin struct. This struct will be used to represent the basic coin in the module.
    struct Coin {}

    // Define the Capabilities struct. This struct will hold the capabilities for minting, freezing, and burning coins.
    struct Capabilities has key {
        burn_cap: coin::BurnCapability<Coin>,
        freeze_cap: coin::FreezeCapability<Coin>,
        mint_cap: coin::MintCapability<Coin>,
    }

    public entry fun init_module(account: &signer) {
        let (burn_cap, freeze_cap, mint_cap)
            = coin::initialize<Coin>(account, string::utf8(b"basic coin"), string::utf8(b"BASIC"), 6);

        let caps = Capabilities { burn_cap, freeze_cap, mint_cap };
        move_to(account, caps);    
    }

    // Define the mint_to function. This function mints a new Coin and deposits it to the specified address.
    public entry fun mint(account: &signer, amount: u64, to: address) acquires Capabilities {
        let addr = signer::address_of(account);
        let caps = borrow_global<Capabilities>(addr);
        let coin = coin::mint<Coin>(amount, &caps.mint_cap);

        coin::deposit<Coin>(to, coin);
    }
}