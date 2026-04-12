package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// Input prompts for a single text value. Returns error if required and empty.
func Input(prompt, placeholder string, required bool) (string, error) {
	var val string
	field := huh.NewInput().
		Title(prompt).
		Placeholder(placeholder).
		Value(&val)
	if required {
		field = field.Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("required")
			}
			return nil
		})
	}
	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

// Password prompts for a password (hidden input).
func Password(prompt string) (string, error) {
	var val string
	field := huh.NewInput().
		Title(prompt).
		EchoMode(huh.EchoModePassword).
		Value(&val)
	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", err
	}
	return val, nil
}

// Select presents a single-choice list. Returns the selected option value.
func Select(title string, options []huh.Option[string]) (string, error) {
	var val string
	field := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&val)
	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", err
	}
	return val, nil
}

// MultiSelect presents a multi-choice list. Returns selected option values.
func MultiSelect(title string, options []huh.Option[string]) ([]string, error) {
	var vals []string
	field := huh.NewMultiSelect[string]().
		Title(title).
		Options(options...).
		Value(&vals)
	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, err
	}
	return vals, nil
}

// Confirm presents a yes/no prompt.
func Confirm(prompt string) (bool, error) {
	var val bool
	field := huh.NewConfirm().
		Title(prompt).
		Value(&val)
	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return false, err
	}
	return val, nil
}

// Text presents a multi-line text input.
func Text(prompt, placeholder string) (string, error) {
	var val string
	field := huh.NewText().
		Title(prompt).
		Placeholder(placeholder).
		Value(&val)
	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}
