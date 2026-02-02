module ember_manager::EmberManager {

    use sui::object;
    use sui::object::UID;
    use sui::tx_context::TxContext;
    use sui::transfer;
    use sui::table;
    use std::option;
    use std::vector;
    use sui::coin::{Coin, USDC}; // adjust import if needed

    /// -------------------------------
    /// Main shared manager object
    /// -------------------------------
    public struct EmberManager has key {
        id: UID,
        loans: table::Table<u64, Loan>,
        next_id: u64,
    }

    /// Initialize and share the manager
    fun init(ctx: &mut TxContext) {
        let manager = EmberManager {
            id: object::new(ctx),
            loans: table::new<u64, Loan>(ctx),
            next_id: 0,
        };

        transfer::share_object(manager);
    }

    /// -------------------------------
    /// Loan (stored inside table)
    /// -------------------------------
    public struct Loan has store {
        amount: u64,
        ltv: u8,
        borrower: address,
        lender: option::Option<address>,
        time: u64,
        htlc_address: vector<u8>,
        state: u8, // 0 = pending, 1 = fulfilled
        hash_secrets_borrower: vector<vector<u8>>,
        hash_secrets_lender: vector<vector<u8>>,
        hashLoanActiveSecret: vector<u8>,
        hashLoanRepaymentSecret: vector<u8>,
    }

    /// -------------------------------
    /// Create loan → stored in table
    /// -------------------------------
    public fun create_loan(
        manager: &mut EmberManager,
        borrower: address,
        amount: u64,
        ltv: u8,
        time: u64,
        hash_secrets_borrower: vector<vector<u8>>,
        hashLoanActiveSecret: vector<u8>,
        hashloanRepaymentSecret: vector<u8>,
    ): u64 {
        assert!(vector::length(&hash_secrets_borrower) == 14, 1);

        let loan_id = manager.next_id;
        manager.next_id = loan_id + 1;

        let loan = Loan {
            amount,
            ltv,
            borrower,
            lender: option::none(),
            time,
            htlc_address: vector::empty(),
            state: 0,
            hash_secrets_borrower,
            hash_secrets_lender: vector::empty(),
            hashLoanActiveSecret,
            hashLoanRepaymentSecret,
        };

        table::add(&mut manager.loans, loan_id, loan);
        loan_id
    }

    /// -------------------------------
    /// Fulfill loan by ID
    /// -------------------------------
    public fun fulfill_loan(
    manager: &mut EmberManager,
    loan_id: u64,
    lender: address,
    hash_secrets_lender: vector<vector<u8>>,
    coin: Coin<USDC>,     // <-- lender sends this coin
    ) {
        assert!(vector::length(&hash_secrets_lender) == 14, 2);
        assert!(table::contains(&manager.loans, loan_id), 3);

        let loan = table::borrow_mut(&mut manager.loans, loan_id);

        assert!(loan.state == 0, 4);

        // Ensure lender sends the correct loan amount
        assert!(Coin::value(&coin) == loan.amount, 5);

        // Set lender and state
        loan.lender = option::some(lender);
        loan.hash_secrets_lender = hash_secrets_lender;
        loan.state = 1;

        // Transfer loan amount to borrower
        coin::transfer(coin, loan.borrower);
    }

    /// -------------------------------
    /// Read-only getters
    /// -------------------------------
    fun borrow_loan(manager: &EmberManager, loan_id: u64): &Loan {
        assert!(table::contains(&manager.loans, loan_id), 10);
        table::borrow(&manager.loans, loan_id)
    }

    /// -------------------------------
    /// Public getters (ID-based)
    /// -------------------------------
    public fun get_amount(manager: &EmberManager, loan_id: u64): u64 {
        borrow_loan(manager, loan_id).amount
    }

    public fun get_ltv(manager: &EmberManager, loan_id: u64): u8 {
        borrow_loan(manager, loan_id).ltv
    }

    public fun get_state(manager: &EmberManager, loan_id: u64): u8 {
        borrow_loan(manager, loan_id).state
    }

    public fun get_borrower(manager: &EmberManager, loan_id: u64): address {
        borrow_loan(manager, loan_id).borrower
    }

    public fun get_hash_secrets_borrower(
        manager: &EmberManager,
        loan_id: u64
    ): &vector<vector<u8>> {
        &borrow_loan(manager, loan_id).hash_secrets_borrower
    }

    public fun get_hash_secrets_lender(
        manager: &EmberManager,
        loan_id: u64
    ): &vector<vector<u8>> {
        &borrow_loan(manager, loan_id).hash_secrets_lender
    }

    // -------------------------------
    // New getters for new secrets
    // -------------------------------
    public fun get_loan_active_secret(
        manager: &EmberManager,
        loan_id: u64
    ): &vector<u8> {
        &borrow_loan(manager, loan_id).loanActiveSecret
    }

    public fun get_loan_repayment_secret(
        manager: &EmberManager,
        loan_id: u64
    ): &vector<u8> {
        &borrow_loan(manager, loan_id).loanRepaymentSecret
    }
}
