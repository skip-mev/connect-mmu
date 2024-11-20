package validator

type Option func(*Validator)

func WithCMCKey(key string) Option {
	return func(v *Validator) {
		v.cmcAPIKey = key
	}
}
