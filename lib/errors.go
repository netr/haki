package lib

import (
	"errors"
	"fmt"
)

// Plugin configuration errors
var (
	ErrPluginIdentifierRequired  = errors.New("plugin identifier is required")
	ErrPromptPathRequired        = errors.New("prompt path is required")
	ErrGenerationModeRequired    = errors.New("generation mode is required")
	ErrCardTypeModelNameRequired = errors.New("card type model name is required")
	ErrFieldMappingRequired      = errors.New("at least one field mapping is required")
	ErrOutputDirectoryRequired   = errors.New("output directory is required")
	ErrSingleCardModeViolation   = errors.New("plugin configured for single card but received multiple cards")
	ErrInvalidTTSService         = errors.New("invalid TTS service implementation")
	ErrInvalidImageGenService    = errors.New("invalid ImageGen service implementation")
)

// Error types that need formatting
type ErrInvalidGenerationMode struct {
	Mode GenerationMode
}

func (e *ErrInvalidGenerationMode) Error() string {
	return fmt.Sprintf("invalid generation mode: %s", e.Mode)
}

type ErrUnknownServiceType struct {
	ServiceType string
}

func (e *ErrUnknownServiceType) Error() string {
	return fmt.Sprintf("unknown service type: %s", e.ServiceType)
}

type ErrRequiredServiceNotProvided struct {
	ServiceType ServiceType
}

func (e *ErrRequiredServiceNotProvided) Error() string {
	return fmt.Sprintf("required service %s not provided", e.ServiceType)
}
