package bike

import "testing"

func TestBikeError_GivenBikeErrorWhenErrorThenReturnErrorMessage(t *testing.T) {
	// Given
	bikeError := Error{
		messageError: "message",
		errorCode:    ComponentConstructorNull,
	}
	// When
	currentMessage := bikeError.Error()
	// Then
	if currentMessage != bikeError.messageError {
		t.Errorf("Error must return expected value")
	}
}

func TestBikeError_GivenBikeErrorWhenErrorCodeThenReturnErrorCode(t *testing.T) {
	// Given
	bikeError := Error{
		messageError: "message",
		errorCode:    ComponentConstructorNull,
	}
	// When
	currentError := bikeError.ErrorCode()
	// Then
	if currentError != ComponentConstructorNull {
		t.Errorf("ErrorCode must return expected value")
	}
}
