package bldr_manifest_world

import (
	"github.com/s4wave/spacewave/db/bucket"
)

// StartupManifestSkipError describes a startup manifest candidate that was skipped.
type StartupManifestSkipError struct {
	// ObjectKey is the skipped candidate object key.
	ObjectKey string
	// ObjectRef is the object or manifest ref involved in the skip, if known.
	ObjectRef *bucket.ObjectRef
	// Err is the underlying availability error.
	Err error
}

// Error returns the startup manifest skip message.
func (e *StartupManifestSkipError) Error() string {
	// A nil error has no message.
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return "startup manifest candidate[" + e.ObjectKey + "] unavailable"
	}

	// Describe the candidate with its ref, bucket, and root.
	msg := "startup manifest candidate[" + e.ObjectKey + "]"
	if e.ObjectRef != nil {
		if !e.ObjectRef.GetEmpty() {
			msg += " ref=" + e.ObjectRef.MarshalString()
		}
		if bucketID := e.ObjectRef.GetBucketId(); bucketID != "" {
			msg += " bucket=" + bucketID
		}
		if rootRef := e.ObjectRef.GetRootRef(); !rootRef.GetEmpty() {
			msg += " root=" + rootRef.MarshalString()
		}
	}
	return msg + ": " + e.Err.Error()
}

// Unwrap returns the underlying skip cause.
func (e *StartupManifestSkipError) Unwrap() error {
	// A nil error unwraps to nil.
	if e == nil {
		return nil
	}
	return e.Err
}

func newStartupManifestSkipError(objKey string, objRef *bucket.ObjectRef, err error) *StartupManifestSkipError {
	return &StartupManifestSkipError{
		ObjectKey: objKey,
		ObjectRef: objRef.Clone(),
		Err:       err,
	}
}
