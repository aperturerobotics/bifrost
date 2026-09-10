package pairing

import "github.com/pkg/errors"

// SameAccount compares logical account identity within the configured provider.
// Storage participants and Session credentials can differ between replicas.
func SameAccount(left, right *AccountOffer) bool {
	return left.GetAccountId() != "" && left.GetAccountId() == right.GetAccountId() && SessionProviderID(left) == SessionProviderID(right) && left.GetProviderEndpoint() == right.GetProviderEndpoint()
}

// ValidateAccounts requires complete, unmodified account descriptions before selection.
func (c *AccountChoice) ValidateAccounts() error {
	if c.GetOfferedAccount() == nil {
		return errors.New("pairing source did not offer an account")
	}
	for _, offer := range []*AccountOffer{c.GetOfferedAccount(), c.GetReceivingAccount()} {
		if offer == nil {
			continue
		}
		if offer.GetAccountId() == "" || offer.GetOperationId() == "" || offer.GetSelectionContext() != "" {
			return errors.New("pairing account description is incomplete or already selected")
		}
	}
	return nil
}

// Validate checks that the chosen relationship applies to the exchanged accounts.
func (c *AccountChoice) Validate() error {
	// Account identity is required independently of the selected operation.
	if err := c.ValidateAccounts(); err != nil {
		return err
	}

	// Reusing an account or arriving from Home never merges or removes another account.
	switch c.GetOutcome() {
	case AccountOutcome_AccountOutcome_SIGN_IN_OFFERED:
		return nil
	case AccountOutcome_AccountOutcome_SIGN_IN_RECEIVING:
		if c.GetReceivingAccount() != nil {
			return nil
		}
	case AccountOutcome_AccountOutcome_MERGE_INTO_OFFERED, AccountOutcome_AccountOutcome_MERGE_INTO_RECEIVING:
		if c.GetReceivingAccount() != nil && !SameAccount(c.GetOfferedAccount(), c.GetReceivingAccount()) {
			return nil
		}
	}
	return errors.New("select an account outcome supported by both clients")
}

// Merging reports whether the selected account absorbs the other account.
func (c *AccountChoice) Merging() bool {
	return c.GetOutcome() == AccountOutcome_AccountOutcome_MERGE_INTO_OFFERED || c.GetOutcome() == AccountOutcome_AccountOutcome_MERGE_INTO_RECEIVING
}
