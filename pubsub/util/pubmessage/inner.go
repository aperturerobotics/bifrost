package pubmessage

// Validate checks the inner data.
func (i *PubMessageInner) Validate() error {
	if i.GetChannel() == "" {
		return ErrInvalidChannelID
	}
	if err := i.GetTimestamp().Validate(true); err != nil {
		return err
	}
	return nil
}
