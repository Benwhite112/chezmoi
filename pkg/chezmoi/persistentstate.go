package chezmoi

var (
	// ConfigStateBucket is the bucket for recording the config state.
	ConfigStateBucket = []byte("configState")

	// EntryStateBucket is the bucket for recording the entry states.
	EntryStateBucket = []byte("entryState")

	// gitRepoExternalState is the bucket for recording the state of commands
	// that modify directories.
	gitRepoExternalState = []byte("gitRepoExternalState")

	// scriptStateBucket is the bucket for recording the state of run once
	// scripts.
	scriptStateBucket = []byte("scriptState")

	stateFormat = formatJSON{}
)

// A PersistentState is a persistent state.
type PersistentState interface {
	Close() error
	CopyTo(s PersistentState) error
	Data() (any, error)
	Delete(bucket, key []byte) error
	DeleteBucket(bucket []byte) error
	ForEach(bucket []byte, fn func(k, v []byte) error) error
	Get(bucket, key []byte) ([]byte, error)
	Set(bucket, key, value []byte) error
}

// PersistentStateBucketData returns the state data in bucket in s.
func PersistentStateBucketData(s PersistentState, bucket []byte) (map[string]any, error) {
	result := make(map[string]any)
	if err := s.ForEach(bucket, func(k, v []byte) error {
		var value map[string]any
		if err := stateFormat.Unmarshal(v, &value); err != nil {
			return err
		}
		result[string(k)] = value
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// PersistentStateData returns the structured data in s.
func PersistentStateData(s PersistentState) (any, error) {
	configStateData, err := PersistentStateBucketData(s, ConfigStateBucket)
	if err != nil {
		return nil, err
	}
	entryStateData, err := PersistentStateBucketData(s, EntryStateBucket)
	if err != nil {
		return nil, err
	}
	gitRepoExternalData, err := PersistentStateBucketData(s, gitRepoExternalState)
	if err != nil {
		return nil, err
	}
	scriptStateData, err := PersistentStateBucketData(s, scriptStateBucket)
	if err != nil {
		return nil, err
	}
	return struct {
		ConfigState         any `json:"configState" yaml:"configState"`
		EntryState          any `json:"entryState" yaml:"entryState"`
		GitRepoExternalData any `json:"gitRepoExternalState" yaml:"gitRepoExternalState"`
		ScriptState         any `json:"scriptState" yaml:"scriptState"`
	}{
		ConfigState:         configStateData,
		EntryState:          entryStateData,
		GitRepoExternalData: gitRepoExternalData,
		ScriptState:         scriptStateData,
	}, nil
}

// persistentStateGet gets the value associated with key in bucket in s, if it exists.
func persistentStateGet(s PersistentState, bucket, key []byte, value any) (bool, error) {
	data, err := s.Get(bucket, key)
	if err != nil {
		return false, err
	}
	if data == nil {
		return false, nil
	}
	if err := stateFormat.Unmarshal(data, value); err != nil {
		return false, err
	}
	return true, nil
}

// persistentStateSet sets the value associated with key in bucket in s.
func persistentStateSet(s PersistentState, bucket, key []byte, value any) error {
	data, err := stateFormat.Marshal(value)
	if err != nil {
		return err
	}
	return s.Set(bucket, key, data)
}
