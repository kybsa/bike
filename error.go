package bike

// ErrorCode type to enum error codes
type ErrorCode uint8

const (
	// DependencyByIDNotFound error code when a dependency not found by id
	DependencyByIDNotFound ErrorCode = 0
	// DependencyByTypeNotFound error code when a dependency not found by type
	DependencyByTypeNotFound ErrorCode = 1
	// InvalidScope error code when a invalid scope
	InvalidScope ErrorCode = 2
	// InvalidNumArgOnPostConstruct error code when PostConstruct method has arguments
	InvalidNumArgOnPostConstruct ErrorCode = 4
	// InvalidNumArgOnDestroy error code when Destroy method has arguments
	InvalidNumArgOnDestroy ErrorCode = 5
	// InvalidNumberOfReturnValuesOnConstructor error when constructor return invalid number of values
	InvalidNumberOfReturnValuesOnConstructor ErrorCode = 6
	// ConstructorReturnNoPointerValue error when constructor method return a no pointer value
	ConstructorReturnNoPointerValue ErrorCode = 7
	// ComponentConstructorNull error when Constructor property is null
	ComponentConstructorNull ErrorCode = 8
	// ConstructorReturnNotNilError return not nil error
	ConstructorReturnNotNilError ErrorCode = 9
	// ConstructorLastReturnValueIsNotError error when last return isn't type error
	ConstructorLastReturnValueIsNotError ErrorCode = 10
	// PostStartWithScopeDifferentToSingleton error when a component has PostStart and Scope different to Singleton
	PostStartWithScopeDifferentToSingleton ErrorCode = 11
	// PostConstructReturnError error when a PostConstruct return a error
	PostConstructReturnError ErrorCode = 12
	// DuplicateScope error when a scope exist
	DuplicateScope ErrorCode = 13
)

// Error struct with error info
type Error struct {
	messageError string
	errorCode    ErrorCode
}

// Error return message about error
func (_self *Error) Error() string {
	return _self.messageError
}

// ErrorCode return error code about error
func (_self *Error) ErrorCode() ErrorCode {
	return _self.errorCode
}
