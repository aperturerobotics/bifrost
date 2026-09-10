package account_settings

import (
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/provider"
)

// FindAccountMigration returns the durable source authorization for returning Sessions.
func (s *AccountSettings) FindAccountMigration(operationID string) *provider.AccountTransition {
	for _, migration := range s.GetAcceptedMigrations() {
		if migration.GetOperationId() == operationID {
			return migration
		}
	}
	return nil
}

// acceptAccountMigration permits one immutable authorization for each source account.
func (s *AccountSettings) acceptAccountMigration(transition *provider.AccountTransition) error {
	if err := transition.Validate(); err != nil {
		return err
	}
	for _, current := range s.GetAcceptedMigrations() {
		if current.GetOperationId() == transition.GetOperationId() || current.GetSource().EqualVT(transition.GetSource()) {
			if !current.EqualVT(transition) {
				return errors.New("account migration authorization is already fixed")
			}
			return nil
		}
	}
	s.AcceptedMigrations = append(s.AcceptedMigrations, transition.CloneVT())
	return nil
}

// commitAccountTransition prevents retries or old replicas from selecting another account.
func (s *AccountSettings) commitAccountTransition(transition *provider.AccountTransition) error {
	if err := transition.Validate(); err != nil {
		return err
	}
	if s.GetTransition() != nil && !s.GetTransition().EqualVT(transition) {
		return errors.New("this account already has a committed destination")
	}
	s.Transition = transition.CloneVT()
	return nil
}
