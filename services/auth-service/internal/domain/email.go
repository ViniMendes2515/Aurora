package domain

type Email struct {
	value string
}

// NewEmail valida e constrói um Email. Retorna ErrInvalidEmail se inválido.
func NewEmail(raw string) (Email, error) {
	if raw == "" {
		return Email{}, ErrInvalidEmail
	}

	if len(raw) < 3 {
		return Email{}, ErrInvalidEmail
	}

	hasAt := false
	hasDot := false

	for i, c := range raw {
		if c == '@' {
			hasAt = true
		}
		if hasAt && c == '.' && i > 0 {
			hasDot = true
		}
	}

	if !hasAt || !hasDot {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: raw}, nil
}

func (e Email) String() string { return e.value }
